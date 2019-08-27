package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"

	. "github.com/andytruong/redes-writer"
)

func main() {
	cnfFile := flag.String("c", "", "")
	flag.Parse()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	cnf, err := NewConfig(*cnfFile)
	if err != nil {
		logrus.WithError(err).Panic("can not read config file")
	}

	processor, queue, errCh, err := Run(ctx, cnf)
	if err != nil {
		logrus.WithError(err).Panic("startup error")
	}

	go func() {
		defer processor.Close()
		panic(<-errCh)
	}()

	http.HandleFunc("/stats", getStatsHandler(processor, queue))
	logrus.
		WithField("port", cnf.Admin.Url).
		Println("es-writer admin ready")

	logrus.
		WithError(http.ListenAndServe(cnf.Admin.Url, nil)).
		Panic()
}

func getStatsHandler(processor *elastic.BulkProcessor, queue Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		type Stats struct {
			Processor      elastic.BulkProcessorStats `json:"processor"`
			QueueName      string                     `json:"queueName"`
			QueueTotalItem int64                      `json:"queueTotalItem"`
		}

		stats := Stats{
			Processor:      processor.Stats(),
			QueueName:      queue.Name(),
			QueueTotalItem: queue.CountItems(),
		}

		if stats, err := json.Marshal(stats); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_, _ = fmt.Fprintln(w, `{"error": "failed to parse statistics."}`)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write(stats)
		}
	}
}
