package redes_writer

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	var cnf Config
	assert := assert.New(t)
	testConfName := "testConfig.yaml"
	testFile := []byte(`admin:
  url: "0.0.0.0:8484"

redis:
  url: "redis://redis:6379?ssl=false"
  queueName: "es-writer"

elasticsearch:
  url: "http://elasticsearch:9200/?sniff=false"

listener:
  bufferSize: 500
  flushInterval: 1s # for faster CI test running
`)

	err := ioutil.WriteFile(testConfName, testFile, 0755)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	defer os.Remove(testConfName)

	err = fileConfig(&cnf, testConfName)
	if nil != err {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal("0.0.0.0:8484", cnf.Admin.Url)
	assert.Equal("redis://redis:6379?ssl=false", cnf.Redis.Url)
	assert.Equal("es-writer", cnf.Redis.QueueName)
	assert.Equal("http://elasticsearch:9200/?sniff=false", cnf.ElasticSearch.Url)
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

	assert.Equal("env:8484", cnf.Admin.Url)
	assert.Equal("redis://envredis:6379?ssl=true", cnf.Redis.Url)
	assert.Equal("env-es-writer", cnf.Redis.QueueName)
	assert.Equal("http://envasticsearch:9200/?sniff=true", cnf.ElasticSearch.Url)
	assert.Equal(600, cnf.Listener.BufferSize)
	assert.Equal(2*time.Second, cnf.Listener.FlushInterval)
}
