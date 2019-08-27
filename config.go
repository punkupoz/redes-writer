package redes_writer

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// configuration required to run services in interface.go
type Config struct {
	Admin struct {
		Url string `yaml:"url"`
	} `yaml:"admin"`
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

// NewConfig return configuration required to run services in interface.go
func NewConfig(cnfPath string) (*Config, error) {
	file, err := ioutil.ReadFile(cnfPath)
	if err != nil {
		return nil, err
	}

	return setConfigFromBytes(file)
}

// setConfigFromBytes receive a pointer to config and array of bytes of configuration file
// this function modify value in config pointer
func setConfigFromBytes(b []byte) (*Config, error) {
	cnf := &Config{}
	expandedInput := os.ExpandEnv(string(b))

	err := yaml.Unmarshal([]byte(expandedInput), cnf)
	if err != nil {
		return cnf, err
	}

	// setConfigFromEnv receive a pointer to config and prefix of environment variable
	// this function modify value in config pointer
	err = envconfig.Process("REDES_WRITER", cnf)
	if err != nil {
		return nil, err
	}

	return cnf, nil
}
