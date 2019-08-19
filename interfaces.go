package redes_writer

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"gopkg.in/yaml.v2"
)

type (
	Queue interface {
		// to write data to ES, client should write data into redis-queue
		// instead of direct writing.
		// for schema of request, ref Request
		Write(payload ...interface{}) error

		// es-writer's listener watches this queue to process the bulk-able requests.
		Listen(ctx context.Context, errCh chan error) chan string

		// Queue for each es-writer should be have unique name.
		Name() string
	}

	Listener interface {
		// entry point to start the es-writer
		// use ctx to cancel the process.
		Run(ctx context.Context, errCh chan error, q Queue, writer Writer) error
	}

	// from  bulk-able request, send to ElasticServer
	// make this an interface, so that we can mock for unit testing without
	// real elastic-search server.
	Writer func(req *Request) error

	// for the services above working, we need configuration
	Config struct {
		Redis struct {
			Url       string `yaml:"url"`
			QueueName string `yaml:"queueName"`
		} `yaml:"redis"`
		Listener struct {
			BufferSize    int           `yaml:"bufferSize"`
			FlushInterval time.Duration `yaml:"flushInterval"`
		} `yaml:"listener"`
		ElasticSearch struct {
			Url string `yaml:"url"`
		} `yaml:"elasticsearch"`
	}
)

func NewQueue(client *redis.Client, name string) (Queue, error) {
	return newQueue(client, name)
}

func NewListener() Listener {
	return newListener()
}

func NewWriter(ctx context.Context, client *elastic.Client, cnf *Config) (Writer, error) {
	processor, err := client.BulkProcessor().
		Name("es-writer").
		BulkSize(cnf.Listener.BufferSize).
		FlushInterval(cnf.Listener.FlushInterval).
		// Workers(5)                TODO: Learn this feature
		// RetryItemStatusCodes(400) TODO: Learn this feature
		Stats(true).
		Do(ctx)

	if nil != err {
		return nil, err
	}

	return func(req *Request) error {
		if nil != req {
			processor.Add(*req)
		}

		return nil
	}, nil
}

func NewConfig(cnfPath string) (*Config, error) {
	yamlBytes, err := ioutil.ReadFile(cnfPath)
	expandedInput := os.ExpandEnv(string(yamlBytes))

	if nil != err {
		return nil, err
	}

	cnf := &Config{}
	err = yaml.Unmarshal([]byte(expandedInput), cnf)
	if nil != err {
		return nil, err
	}

	return cnf, nil
}

func Run(ctx context.Context, cnfPath string) (chan error, error) {
	cnf, err := NewConfig(cnfPath)
	if nil != err {
		return nil, err
	}

	cElasticSearch, err := newElasticSearchClient(cnf.ElasticSearch.Url)
	if nil != err {
		return nil, err
	}

	cRedis := newRedisClient(cnf.Redis.Url)
	queue, err := NewQueue(cRedis, cnf.Redis.QueueName)
	if nil != err {
		return nil, err
	}

	writer, err := NewWriter(ctx, cElasticSearch, cnf)
	if nil != err {
		return nil, err
	}

	return run(ctx, queue, writer)
}

func run(ctx context.Context, queue Queue, writer Writer) (chan error, error) {
	errCh := make(chan error, 1)
	err := NewListener().Run(ctx, errCh, queue, writer)
	if nil != err {
		return nil, err
	}

	return errCh, nil
}
