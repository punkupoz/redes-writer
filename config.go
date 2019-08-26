package redes_writer

import (
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
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

	err := fileConfig(cnf, cnfPath)
	if err != nil {
		return nil, err
	}

	err = envConfig(cnf, "REDES_WRITER")
	if err != nil {
		return nil, err
	}

	return cnf, nil
}

// fileConfig receive a pointer to config and path to config file path
// this function modify value in config pointer
func fileConfig(cnf *Config, cnfPath string) error {
	file, err := ioutil.ReadFile(cnfPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, cnf)
	if err != nil {
		return err
	}
	return nil
}

// envConfig receive a pointer to config and prefix of environment variable
// this function modify value in config pointer
func envConfig(cnf *Config, prefix string) error {
	err := envconfig.Process(prefix, cnf)
	if err != nil {
		return err
	}
	return nil
}