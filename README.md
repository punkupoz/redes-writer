Elastic Search writer [![Build Status](https://travis-ci.org/andytruong/es-writer.svg?branch=master)](https://travis-ci.org/andytruong/es-writer)
====

## Problems

It's very easy to have conflict when we have multiple services writing data into same Elastic Search server.

To avoid this problem, the service should publish message to a certain instead of writing to ES directly. So 
that we can have ES-Writer, a single actor that connects with ElasticSearch.

By this convention, the services doesn't need to know credentials of Elastic Search server.

It's also easy to create cluster ElasticSearch servers without magic.

## Usage

Start the worker

    es-writer -c /path/to/config.yaml

Start new requests

    redis-cli > RPUSH $queueName $bulkableRequest1
              > RPUSH $queueName $bulkableRequest2 $bulkableRequest3

### Test local

    docker run -d -p 9200:9200                --rm --name es       elasticsearch:5-alpine
    docker run -d -p 5672:5672 -p 15672:15672 --rm --name rabbitmq rabbitmq:3-management
    go test -race -v ./...
