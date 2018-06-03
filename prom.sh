#!/bin/sh

export GOMAXPROCS=4
~/src/blts/prometheus-2.2.1.linux-amd64/prometheus \
    --config.file=prom/prometheus.yml \
    --web.external-url=http://127.0.0.1:9090 \
    --web.listen-address=127.0.0.1:9090 \
    --log.level=debug
