#!/bin/sh

set -e

export GOMAXPROCS=4

BACKENDS=""
for i in $(seq 0 9); do
    (s/s --port 800$i 2>&1 | sed -e "s/^/s:800$i: /") &
    BACKENDS="$BACKENDS :800$i"
done

BACKENDS=$(echo $BACKENDS | tr ' ' ',')
(lb/lb --backends $BACKENDS 2>&1 | sed -e "s/^/lb: /") &

rm -rf data
mkdir -p data
TARGETS="$BACKENDS,:9001"
(c/c --targets $TARGETS 2>&1 | sed -e "s/^/c: /") &

wait
