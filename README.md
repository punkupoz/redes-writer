Elastic Search writer [![Build Status](https://travis-ci.org/andytruong/redes-writer.svg?branch=7.x)](https://travis-ci.org/andytruong/redes-writer) [![](https://images.microbadger.com/badges/image/andytruong/redes-writer.svg)](https://microbadger.com/images/andytruong/redes-writer "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/andytruong/redes-writer.svg)](https://microbadger.com/images/andytruong/redes-writer "Get your own version badge on microbadger.com")
====

## Problems

It's very easy to have conflict when we have multiple services writing data into same Elastic Search server.

To avoid this problem, the service should publish message to a certain instead of writing to ES directly. So 
that we can have ES-Writer, a single actor that connects with ElasticSearch.

By this convention, the services doesn't need to know credentials of Elastic Search server.

It's also easy to create cluster ElasticSearch servers without magic.

## Usage

Manual testing

    git clone https://github.com/andytruong/redes-writer.git
    cd redes-writer
    docker-compose up -d
    
    # queue bulkable request
    # -------
    docker-compose exec redis sh -c redis-cli
    127.0.0.1:6379> RPUSH es-writer '{"type": "index","index": {"index": "lr","type":  "enrolment","id":    "123","routing": "456","doc": {"field1" : "value1"}}}'
    127.0.0.1:6379> RPUSH es-writer '{"type": "update", "update": { "index": "lr", "type":  "enrolment", "id":    "123", "routing": "456", "doc": { "field2" : "value2" }}}'
    127.0.0.1:6379> PUBLISH es-writer-pubsub 1
    
    # check ES server for expeting result
    # -------
    docker-compose exec elasticsearch curl localhost:9200/lr/enrolment/123
    
    # Clean up
    # -------
    docker-compose down

Start servers

    docker run -d -p 6379:6379 --rm --name=hi-redis redis:5.0-alpine
    docker run -d -p 9200:9200 --rm --name=hi-es7 -e "discovery.type=single-node"  docker.elastic.co/elasticsearch/elasticsearch:7.3.0

Start the worker

    es-writer -c /path/to/config.yaml

Start new requests

    redis-cli > RPUSH $queueName $bulkableRequest1
              > RPUSH $queueName $bulkableRequest2 $bulkableRequest3

Test
    
    go test -race -v ./...
    
Docker compose with Datadog

    DD_API_KEY=<YOUR_DD_API_KEY> docker-compose -f docker-compose.yml -f datadog-compose.yml up
    
