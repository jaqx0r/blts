#!/bin/sh

rm -rf data # Default local storage.
~/src/prometheus/prometheus/prometheus -config.file=prom/prometheus.yml -alertmanager.url=http://localhost:9093
