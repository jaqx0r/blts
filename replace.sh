#!/bin/sh

export GOMAXPROCS=4
renice +10 $$

kill $1
s/s --port 8009 --faily
