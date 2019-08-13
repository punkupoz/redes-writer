package redes_writer

import (
	"context"
	"crypto/tls"
	"net/url"
	"time"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/olivere/elastic/v7/config"
	"github.com/sirupsen/logrus"
)

func newRedisClient(redisUrl string) *redis.Client {
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

func newElasticSearchClient(url string) (*elastic.Client, error) {
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

func readTimeout(source chan string) string {
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()

	select {
	case msg := <-source:
		return msg

	case <-ctx.Done():
		return ""
	}
}
