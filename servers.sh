#!/bin/sh

set -e

export GOMAXPROCS=4

BACKENDS=""
for i in $(seq 0 9); do
    s/s --port 800$i &
    BACKENDS="$BACKENDS :800$i"
done

BACKENDS=$(echo $BACKENDS | tr ' ' ',')
lb/lb --backends $BACKENDS &

mkdir -p data
TARGETS="$BACKENDS,:8080"
c/c --targets $TARGETS &

wait
