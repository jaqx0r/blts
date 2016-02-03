#!/bin/sh

export GOMAXPROCS=4
~/go/src/github.com/prometheus/prometheus/prometheus \
    -config.file=prom/prometheus.yml \
    -alertmanager.url=http://127.0.0.1:9093 \
    -storage.local.path=./prom-data \
    -web.external-url=http://127.0.0.1:9090 \
    -web.listen-address=127.0.0.1:9090 \
    -log.level=debug
