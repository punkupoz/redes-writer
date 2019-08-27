package redes_writer

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	a := assert.New(t)

	envPrefix := "REDES_WRITER"
	_ = os.Setenv(envPrefix+"_ADMIN_URL", "envenv:8484")
	_ = os.Setenv(envPrefix+"_REDIS_URL", "redis://rere:6379?ssl=false")
	_ = os.Setenv(envPrefix+"_ELASTICSEARCH_URL", "http://searchelast:9200/?sniff=true")

	cnf, err := NewConfig("config.sample.yaml")
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	a.Equal("envenv:8484", cnf.Admin.Url)
	a.Equal("redis://rere:6379?ssl=false", cnf.Redis.Url)
	a.Equal("es-writer", cnf.Redis.QueueName)
	a.Equal("http://searchelast:9200/?sniff=true", cnf.ElasticSearch.Url)
	a.Equal(500, cnf.Listener.BufferSize)
	a.Equal(1*time.Second, cnf.Listener.FlushInterval)
}

func TestEnvOverride(t *testing.T) {
	envPrefix := "REDES_WRITER"
	_ = os.Setenv(envPrefix+"_ADMIN_URL", "env:8484")
	_ = os.Setenv(envPrefix+"_REDIS_URL", "redis://envredis:6379?ssl=true")
	_ = os.Setenv(envPrefix+"_REDIS_QUEUENAME", "env-es-writer")
	_ = os.Setenv(envPrefix+"_ELASTICSEARCH_URL", "http://envasticsearch:9200/?sniff=true")
	_ = os.Setenv(envPrefix+"_LISTENER_BUFFERSIZE", "600")
	_ = os.Setenv(envPrefix+"_LISTENER_FLUSHINTERVAL", "2s")
	testFile := []byte(`
# expecting confiugration will be parsed from environment variables.
admin:         {}
redis:         {}
elasticsearch: {}
listener:      {}
`)

	cnf, err := setConfigFromBytes(testFile)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	a := assert.New(t)
	a.Equal("env:8484", cnf.Admin.Url)
	a.Equal("redis://envredis:6379?ssl=true", cnf.Redis.Url)
	a.Equal("env-es-writer", cnf.Redis.QueueName)
	a.Equal("http://envasticsearch:9200/?sniff=true", cnf.ElasticSearch.Url)
	a.Equal(600, cnf.Listener.BufferSize)
	a.Equal(2*time.Second, cnf.Listener.FlushInterval)
}
