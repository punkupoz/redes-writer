package redes_writer

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	envPrefix := "REDES_WRITER"

	_ = os.Setenv(envPrefix+"_ADMIN_URL", "envenv:8484")
	_ = os.Setenv(envPrefix+"_REDIS_URL", "redis://rere:6379?ssl=false")
	_ = os.Setenv(envPrefix+"_ELASTICSEARCH_URL", "http://searchelast:9200/?sniff=true")

	cnf, err := NewConfig("config.sample.yaml")
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	// redis connection
	assert.Equal("envenv:8484", cnf.Admin.Url)
	assert.Equal("redis://rere:6379?ssl=false", cnf.Redis.Url)

	// ES connection
	assert.Equal("es-writer", cnf.Redis.QueueName)
	assert.Equal("http://searchelast:9200/?sniff=true", cnf.ElasticSearch.Url)

	// Listener
	assert.Equal(500, cnf.Listener.BufferSize)
	assert.Equal(1*time.Second, cnf.Listener.FlushInterval)
}

func TestEnvOverride(t *testing.T) {
	var cnf Config
	assert := assert.New(t)
	testFile := []byte(`admin:
  url: ${ADMIN_URL}

redis:
  url: "redis://redis:6379?ssl=false"
  queueName: "es-writer"

elasticsearch:
  url: "http://elasticsearch:9200/?sniff=false"

listener:
  bufferSize: 500
  flushInterval: 1s # for faster CI test running
`)

	os.Setenv("ADMIN_URL", "0.0.0.0:8484")

	err := setConfigFromBytes(&cnf, testFile)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	// redis connection
	assert.Equal("0.0.0.0:8484", cnf.Admin.Url)
	assert.Equal("redis://redis:6379?ssl=false", cnf.Redis.Url)

	// ES connection
	assert.Equal("es-writer", cnf.Redis.QueueName)
	assert.Equal("http://elasticsearch:9200/?sniff=false", cnf.ElasticSearch.Url)

	// Listener
	assert.Equal(500, cnf.Listener.BufferSize)
	assert.Equal(1*time.Second, cnf.Listener.FlushInterval)

	envPrefix := "TEST_REDES_WRITER"

	_ = os.Setenv(envPrefix+"_ADMIN_URL", "env:8484")
	_ = os.Setenv(envPrefix+"_REDIS_URL", "redis://envredis:6379?ssl=true")
	_ = os.Setenv(envPrefix+"_REDIS_QUEUENAME", "env-es-writer")
	_ = os.Setenv(envPrefix+"_ELASTICSEARCH_URL", "http://envasticsearch:9200/?sniff=true")
	_ = os.Setenv(envPrefix+"_LISTENER_BUFFERSIZE", "600")
	_ = os.Setenv(envPrefix+"_LISTENER_FLUSHINTERVAL", "2s")

	err = envconfig.Process(envPrefix, &cnf)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	// Redis connection
	assert.Equal("env:8484", cnf.Admin.Url)
	assert.Equal("redis://envredis:6379?ssl=true", cnf.Redis.Url)

	// ES connection
	assert.Equal("env-es-writer", cnf.Redis.QueueName)
	assert.Equal("http://envasticsearch:9200/?sniff=true", cnf.ElasticSearch.Url)

	// Listener
	assert.Equal(600, cnf.Listener.BufferSize)
	assert.Equal(2*time.Second, cnf.Listener.FlushInterval)
}