package redes_writer

import (
	"context"

	"github.com/sirupsen/logrus"
)

type listener struct{}

func newListener() Listener {
	return &listener{}
}

func (l *listener) Run(ctx context.Context, errCh chan error, q Queue, writer Writer) error {
	ch := q.Listen(ctx, errCh)

	go func(ctx context.Context) {
		for {
			raw := <-ch
			req, err := fromBytes(raw)
			if err != nil {
				errCh <- err
			}

			err = writer(req)
			if err != nil {
				errCh <- err
			}

			select {
			case <-ctx.Done():
				logrus.Infoln("cancelled 🐰 listening")
			default: // just continue
			}
		}
	}(ctx)

	return nil
}
