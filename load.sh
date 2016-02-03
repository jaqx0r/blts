#!/bin/sh

renice +10 $$

max=10
for i in $(seq 1 $(( 2 * $max )) ); do
    (ab -r -n 100000 -c 1000 http://localhost:9001/hi 2>&1 | sed -e "s/^/$i: /") &
    sleep $(echo $(( $max - $i )) | tr -d '-')
done

wait
