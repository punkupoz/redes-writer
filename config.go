package redes_writer

import (
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"time"
)

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

func envConfig(cnf *Config, prefix string) error {
	err := envconfig.Process(prefix, cnf)
	if err != nil {
		return err
	}
	return nil
}