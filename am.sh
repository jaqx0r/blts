#!/bin/sh

~/go/src/github.com/prometheus/alertmanager/alertmanager -config.file=prom/alertmanager.yml -storage.path=./am-data
