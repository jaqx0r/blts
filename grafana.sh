#!/bin/sh

GRAF=`pwd`/graf

/usr/sbin/grafana-server \
	--homepath=/usr/share/grafana \
	--config=${GRAF}/grafana.ini \
	cfg:default.paths.provisioning=${GRAF}/prov \
	cfg:default.paths.data=${GRAF}/data \
	cfg:default.paths.logs=${GRAF}/logs \
	cfg:default.paths.plugins=/usr/share/grafana/plugins \

