#!/bin/sh

#rm -rf data # Default local storage.
~/go/src/github.com/prometheus/prometheus/prometheus -config.file=prom/prometheus.yml -alertmanager.url=http://127.0.0.1:9093 -storage.local.path=./prom-data -web.external-url=http://127.0.0.1:9090
