package es_writer

import (
	"crypto/tls"
	"net/url"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/olivere/elastic/v7/config"
	"github.com/sirupsen/logrus"
)

func NewRedisClient(redisUrl string) *redis.Client {
	u, err := url.Parse(redisUrl)
	if err != nil {
		panic(err)
	}

	pass, _ := u.User.Password()
	options := &redis.Options{
		Addr:     u.Host,
		Password: pass,
		DB:       0,
	}

	if u.Query().Get("ssl") == "true" {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return redis.NewClient(options)
}

func NewElasticSearchClient(url string) (*elastic.Client, error) {
	cfg, err := config.Parse(url)
	if err != nil {
		logrus.Fatalf("failed to parse URL: %s", err.Error())

		return nil, err
	}

	client, err := elastic.NewClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
