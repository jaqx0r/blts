#!/bin/sh

ab -r -c 1000 -t 3600 http://localhost:9001/hi
