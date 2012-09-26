#!/bin/sh

max=10
for i in $(seq 1 $(( 2 * $max )) ); do
    (ab -n 10000 -c 100 http://localhost:9001/hi 2>&1 | sed -e "s/^/$i: /") &
    sleep $(echo $(( $max - $i )) | tr -d '-')
done

wait
