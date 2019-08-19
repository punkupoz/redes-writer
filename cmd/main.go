package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	. "github.com/andytruong/redes-writer"
)

func main() {
	cnfFile := flag.String("c", "", "")
	flag.Parse()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	errCh, err := Run(ctx, *cnfFile)

	if err != nil {
		logrus.WithError(err).Panic("startup error")
	}

	err = <-errCh
	logrus.WithError(err).Panic("runtime error")
}
