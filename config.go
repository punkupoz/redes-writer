package redes_writer

import (
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"os"
	"time"
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
	cnf := &Config{}
	file, err := ioutil.ReadFile(cnfPath)
	if err != nil {
		return nil, err
	}

	err = setConfigFromBytes(cnf, file)
	if err != nil {
		return nil, err
	}

	err = setConfigFromEnv(cnf, "REDES_WRITER")
	if err != nil {
		return nil, err
	}

	return cnf, nil
}

// setConfigFromBytes receive a pointer to config and array of bytes of configuration file
// this function modify value in config pointer
func setConfigFromBytes(cnf *Config, b []byte) error {
	expandedInput := os.ExpandEnv(string(b))

	err := yaml.Unmarshal([]byte(expandedInput), cnf)
	if err != nil {
		return err
	}

	return nil
}

// setConfigFromEnv receive a pointer to config and prefix of environment variable
// this function modify value in config pointer
func setConfigFromEnv(cnf *Config, prefix string) error {
	err := envconfig.Process(prefix, cnf)
	if err != nil {
		return err
	}
	return nil
}