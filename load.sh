#!/bin/sh

for i in $(seq 1 10); do
    (ab -r -n 100000 -c 100 http://localhost:8080/hi 2>&1 | sed -e "s/^/$i: /") &
    sleep 10
done

wait

