#!/bin/sh

~/go/src/github.com/prometheus/alertmanager/alertmanager -config.file=prom/alertmanager.yml -storage.path=./am-data -web.listen-address=127.0.0.1:9093 -web.external-url=http://127.0.0.1:9093
