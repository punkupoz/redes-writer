package es_writer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type queue struct {
	name    string
	client  *redis.Client
	ps      *redis.PubSub
	timeout time.Duration
}

func (q queue) Name() string {
	return q.name
}

func (q queue) pubsubChanel() string {
	return q.Name() + "-pubsub"
}

func newQueue(client *redis.Client, name string) (*queue, error) {
	q := &queue{
		name:    name,
		client:  client,
		ps:      nil,
		timeout: 3 * time.Second,
	}

	q.ps = client.Subscribe(q.pubsubChanel())

	// wait for pubsub connection successfully connected.
	_, err := q.ps.ReceiveTimeout(time.Second)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func (q queue) Write(payload ...interface{}) error {
	cmd := q.client.RPush(q.Name(), payload...)

	if err := q.ps.Ping(); err != nil {
		return err
	}

	return cmd.Err()
}

func (q queue) Listen(ctx context.Context, errCh chan error) chan string {
	ch := make(chan string)

	defer func() {
		if err := recover(); err != nil {
			errCh <- fmt.Errorf("es-writer is broken: %s", err)
		}
	}()

	go q.loop(ctx, q.sub(ctx, errCh), ch)

	return ch
}

func (q *queue) loop(ctx context.Context, sub chan string, ch chan string) {
	for { // run forever
		for { // process all items in queue
			result, err := q.client.LPop(q.Name()).Result()
			if nil != err {
				if err.Error() != "redis: nil" {
					panic(err)
				}
			}

			// queue is now empty, don't need fetching it again
			if 0 == len(result) {
				break
			}

			ch <- result
		}

		select {
		case <-ctx.Done(): // got cancel signature, stop
			close(ch)
			return

		case <-sub: // wait for signal form ps channel
			continue
		}
	}
}

func (q queue) sub(ctx context.Context, errCh chan error) chan string {
	ch := make(chan string, 1)

	go func() {
		for {
			msg, err := q.ps.ReceiveMessage()
			if nil != err {
				errCh <- err
			}

			// we may have too many signal in ps channel
			// we should send-out only one
			// q.ps.ReceiveTimeout()
			for {
				_, err := q.ps.ReceiveTimeout(time.Second)
				if err != nil {
					if strings.Contains(err.Error(), "i/o timeout") {
						break
					} else {
						errCh <- err
					}
				}
			}

			ch <- msg.Payload
		}
	}()

	return ch
}
