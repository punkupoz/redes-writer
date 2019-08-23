package redes_writer

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func redisUrl() string {
	if env := os.Getenv("REDIS_URL"); "" != env {
		return env
	}

	return "redis://localhost:6379?ssl=false"
}

func esUrl() string {
	if env := os.Getenv("ES_URL"); "" != env {
		return env
	}

	return "http://127.0.0.1:9200/?sniff=false"
}

func TestConfig_Parse(t *testing.T) {
	cnf, err := NewConfig("config.sample.yaml")
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	// redis connection
	assert.Equal(t, "redis://redis:6379?ssl=false", cnf.Redis.Url)
	assert.Equal(t, "es-writer", cnf.Redis.QueueName)

	// ES connection
	assert.Equal(t, "http://elasticsearch:9200/?sniff=false", cnf.ElasticSearch.Url)

	// Listener
	assert.Equal(t, 1*time.Second, cnf.Listener.FlushInterval)
	assert.Equal(t, 500, cnf.Listener.BufferSize)
}

func TestRequest_ToBulkIndex(t *testing.T) {
	raw := []byte(`
		{
			"type": "index",
			"index": {
				"index": "lr",
				"type":  "enrolment",
				"id":    "123",
				"routing": "456",
				"doc": { "field1" : "value1" }
			}
		}
	`)

	req := Request{}

	err := json.Unmarshal(raw, &req)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	a := assert.New(t)
	output, _ := req.Source()

	a.Equal(`{"index":{"_index":"lr","_id":"123","_type":"enrolment","retry_on_conflict":0,"routing":"456"}}`, output[0])
	a.Equal(`{"field1":"value1"}`, output[1])
}

func TestRequest_ToBulkUpdate(t *testing.T) {
	raw := []byte(`
		{
			"type": "update",
			"update": {
				"index": "lr",
				"type":  "enrolment",
				"id":    "123",
				"routing": "456",
				"doc": { "field2" : "value2" }
			}
		}
	`)

	req := Request{}

	err := json.Unmarshal(raw, &req)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	a := assert.New(t)
	output, _ := req.Source()

	a.Contains(output[0], `"update":{"_index":"lr"`)
	a.Contains(output[0], `"_id":"123"`)
	a.Contains(output[0], `"_type":"enrolment"`)
	a.Contains(output[0], `"routing":"456"`)
	a.Equal(`{"doc":{"field2":"value2"}}`, output[1])
}

func TestRequest_ToBulkDelete(t *testing.T) {
	raw := []byte(`
		{
			"type": "delete",
			"delete": {
				"index": "lr",
				"type":  "enrolment",
				"id":    "123",
				"routing": "456"
			}
		}
	`)

	req := Request{}
	err := json.Unmarshal(raw, &req)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	a := assert.New(t)
	output, _ := req.Source()

	a.Equal(`{"delete":{"_index":"lr","_type":"enrolment","_id":"123","routing":"456"}}`, output[0])
}

func TestQueue_Learn(t *testing.T) {
	client := newRedisClient(redisUrl())

	// learn redis API to ping server
	{
		client.FlushAll()
		pong, err := client.Ping().Result()
		if nil != err {
			t.Error(err)
			t.FailNow()
		}

		assert.Equal(t, "PONG", pong)
	}

	ps := client.Subscribe("myPubSub")
	defer ps.Close()

	{ // very basic ping pong
		client.FlushAll()
		client.Publish("myPubSub", "ping")

		msg, _ := ps.ReceiveMessage()
		assert.Equal(t, "ping", msg.Payload)
	}

	{ // read with timeout
		client.FlushAll()
		_, err := ps.ReceiveTimeout(300 * time.Millisecond)
		assert.Contains(t, err.Error(), "i/o timeout")
	}

	{ // rpop from an empty list
		client.FlushAll()
		client.LPush("myList", "one")
		one := client.RPop("myList").Val()
		assert.Equal(t, "one", one)

		_, err := client.RPop("myList").Result()
		assert.Equal(t, "redis: nil", err.Error())
	}
}

func TestQueue_Listen(t *testing.T) {
	client := newRedisClient(redisUrl())
	client.FlushAll()

	queue, _ := NewQueue(client, "myQueue")
	ctx, stop := context.WithCancel(context.TODO())
	ch := queue.Listen(ctx, make(chan error))

	m1 := `{"type": "index","index": {"index": "lr","type":  "enrolment","id":    "123","routing": "456","doc": {"field1" : "value1"}}}`
	m2 := `{"type": "update", "update": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456", "doc": { "field2" : "value2" }}}`
	m3 := `{"type": "delete", "delete": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456"}}`
	err := queue.Write(m1, m2, m3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, m1, <-ch)
	assert.Equal(t, m2, <-ch)
	assert.Equal(t, m3, <-ch)

	stop()
}

func TestQueue_Subscription(t *testing.T) {
	client := newRedisClient(redisUrl())
	client.FlushAll()

	queue, _ := newQueue(client, "myQueue")
	ctx, stop := context.WithCancel(context.TODO())
	sub := queue.sub(ctx, make(chan error, 1))

	defer stop()

	{ // publish & got notification
		_ = client.Publish(queue.pubsubChanel(), "000")
		assert.Equal(t, "000", readTimeout(sub))
	}

	{ // publish & publish again -> both should got notification
		client.Publish(queue.pubsubChanel(), 111)
		assert.Equal(t, "111", readTimeout(sub))

		_ = client.Publish(queue.pubsubChanel(), 222)
		assert.Equal(t, "222", readTimeout(sub))

		// no more message to read.
		assert.Equal(t, "", readTimeout(sub))
	}

	{ // Test that we pub 3 messages, only receive one notification from subscription
		_ = client.Publish(queue.pubsubChanel(), 333)
		_ = client.Publish(queue.pubsubChanel(), 444)

		// we should only receive one notification.
		assert.Equal(t, "333", readTimeout(sub))

		// check again, should receive timeout error because all consumed.
		assert.Equal(t, "", readTimeout(sub))
	}
}

func TestListener_Run(t *testing.T) {
	client := newRedisClient(redisUrl())
	client.FlushAll()
	queue, _ := NewQueue(client, "myQueue")

	l := newListener()
	recorder := []string{}
	ctx, cancel := context.WithCancel(context.TODO())
	wg := sync.WaitGroup{}

	// start listener
	// -------
	err := l.Run(
		ctx,
		make(chan error),
		queue,
		func(req *Request) error {
			defer wg.Done()
			recorder = append(recorder, req.String())

			return nil
		},
	)

	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	// send some requests into queue
	m1 := `{"type": "index","index": {"index": "lr","type":  "enrolment","id":    "123","routing": "456","doc": {"field1" : "value1"}}}`
	m2 := `{"type": "update", "update": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456", "doc": { "field2" : "value2" }}}`
	m3 := `{"type": "delete", "delete": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456"}}`
	wg.Add(3)
	if err := queue.Write(m1, m2, m3); err != nil {
		t.Error(err)
		t.FailNow()
	}

	wg.Wait()

	assert.Contains(t, recorder[0], `{"index":{"_index":"lr","_id":"123","_type":"enrolment","retry_on_conflict":0,"routing":"456"}}`)
	assert.Contains(t, recorder[0], `{"field1":"value1"}`)
	assert.Contains(t, recorder[1], `{"update":{"_index":"lr","_type":"enrolment","_id":"123","routing":"456"}}`)
	assert.Contains(t, recorder[1], `{"doc":{"field2":"value2"}}`)
	assert.Contains(t, recorder[2], `{"delete":{"_index":"lr","_type":"enrolment","_id":"123","routing":"456"}}`)

	cancel()
}

func TestEndToEnd(t *testing.T) {
	ctx, done := context.WithCancel(context.TODO())
	defer done()

	// create services & start the application
	es, _ := newElasticSearchClient(esUrl())

	_, err := es.CreateIndex("lr").Do(ctx)
	if err != nil {
		if !strings.Contains(err.Error(), "esource_already_exists_exception") {
			t.Error(err)
			t.FailNow()
		}
	}

	cnf, _ := NewConfig("config.sample.yaml")
	processor, err := NewProcessor(ctx, es, cnf)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	writer, _ := NewWriter(context.WithValue(ctx, "processor", processor))
	client := newRedisClient(redisUrl())
	client.FlushAll()
	queue, _ := NewQueue(client, "myQueue")
	_, _ = run(ctx, queue, writer)

	// send some requests into queue
	m1 := `{"type": "index","index": {"index": "lr","type":  "enrolment","id":    "123","routing": "456","doc": {"field1" : "value1"}}}`
	m2 := `{"type": "update", "update": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456", "doc": { "field2" : "value2" }}}`

	if err := queue.Write(m1, m2); err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(4 * time.Second)

	// Check stats
	res, err := es.Search("lr").Routing("456").Do(ctx)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, res.TotalHits() > 0)

	doc, _ := res.Hits.Hits[0].Source.MarshalJSON()
	assert.Equal(t, `{"field1":"value1","field2":"value2"}`, string(doc))
}
