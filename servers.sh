#!/bin/sh

BACKENDS=""

for i in $(seq 0 9); do
    s/s --port 800$i &
    BACKENDS="$BACKENDS :800$i"
done

BACKENDS=$(echo $BACKENDS | tr ' ' ',')
lb/lb --backends $BACKENDS &

wait
