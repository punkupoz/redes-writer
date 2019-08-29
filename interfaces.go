package redes_writer

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type (
	Queue interface {
		// to write data to ES, client should write data into redis-queue
		// instead of direct writing.
		// for schema of request, ref Request
		Write(payload ...interface{}) error

		// es-writer's listener watches this queue to process the bulk-able requests.
		Listen(ctx context.Context, mc *metricCollector, errCh chan error) chan string

		// Queue for each es-writer should be have unique name.
		Name() string

		CountItems() int64
	}

	Listener interface {
		// entry point to start the es-writer
		// use ctx to cancel the process.
		Run(ctx context.Context, errCh chan error, q Queue, writer Writer, mc *metricCollector) error
	}

	Metric interface {

	}

	// from  bulk-able request, send to ElasticServer
	// make this an interface, so that we can mock for unit testing without
	// real elastic-search server.
	Writer func(req *Request) error
)

func NewQueue(client *redis.Client, name string) (Queue, error) {
	return newQueue(client, name)
}

func NewListener() Listener {
	return newListener()
}

func NewProcessor(ctx context.Context, client *elastic.Client, cnf *Config) (*elastic.BulkProcessor, error) {
	// should read: https://github.com/olivere/elastic/wiki/BulkProcessor

	return client.BulkProcessor().
		Name("es-writer").
		BulkSize(cnf.Listener.BufferSize).
		FlushInterval(cnf.Listener.FlushInterval).
		Stats(true).
		// Workers(5)                TODO: Learn this feature
		// RetryItemStatusCodes(400) // default: 408, 429, 503, 507
		After(
			func(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
				if err != nil {
					logrus.WithError(err).Errorln("process error")
				}

				for _, rItem := range response.Items {
					for riKey, riValue := range rItem {
						if riValue.Error != nil {
							logrus.
								WithField("key", riKey).
								WithField("type", riValue.Error.Type).
								WithField("phase", riValue.Error.Phase).
								WithField("reason", riValue.Error.Reason).
								Errorf("failed to process item %s", riKey)
						}
					}
				}
			},
		).
		Do(ctx)
}

func NewWriter(ctx context.Context) (Writer, error) {
	processor := ctx.Value("processor").(*elastic.BulkProcessor)

	return func(req *Request) error {
		if nil != req {
			processor.Add(*req)
		}

		return nil
	}, nil
}


func Run(ctx context.Context, cnf *Config, mc *metricCollector) (*elastic.BulkProcessor, Queue, chan error, error) {
	cElasticSearch, err := newElasticSearchClient(cnf.ElasticSearch.Url)
	if nil != err {
		return nil, nil, nil, err
	}

	cRedis := newRedisClient(cnf.Redis.Url)
	queue, err := NewQueue(cRedis, cnf.Redis.QueueName)
	if nil != err {
		return nil, nil, nil, err
	}

	processor, err := NewProcessor(ctx, cElasticSearch, cnf)
	if nil != err {
		return nil, nil, nil, err
	}

	ctx = context.WithValue(ctx, "processor", processor)
	writer, err := NewWriter(ctx)
	if nil != err {
		return nil, nil, nil, err
	}

	errCh, err := run(ctx, queue, writer, mc)

	return processor, queue, errCh, err
}

func run(ctx context.Context, queue Queue, writer Writer, mc *metricCollector) (chan error, error) {
	errCh := make(chan error, 1)
	err := NewListener().Run(ctx, errCh, queue, writer, mc)
	if nil != err {
		return nil, err
	}

	return errCh, nil
}
