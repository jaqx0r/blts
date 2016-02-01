#!/bin/sh

set -e

trap "killall lb; killall s" HUP TERM INT EXIT

export GOMAXPROCS=4

renice +10 $$

BACKENDS=""
for i in $(seq 0 9); do
    (s/s --port 800$i 2>&1 | sed -e "s/^/s:800$i: /") &
    BACKENDS="$BACKENDS :800$i"
done

BACKENDS=$(echo $BACKENDS | tr ' ' ',')
(lb/lb --backends $BACKENDS 2>&1 | sed -e "s/^/lb: /") &

wait
