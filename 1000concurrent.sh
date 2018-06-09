#!/bin/sh

ab -r -c 1000 -t 3600 -n 1000000 http://localhost:9001/hi
