package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	. "github.com/andytruong/es-writer"
)

func main() {
	cnfFile := flag.String("c", "", "")
	flag.Parse()

	ctx := context.Background()
	errCh, err := Run(ctx, *cnfFile)

	if err != nil {
		logrus.WithError(err).Panic("startup error")
	}

	err = <-errCh
	logrus.WithError(err).Panic("runtime error")
}
