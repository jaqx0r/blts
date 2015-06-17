#!/bin/sh

ab -r -c 10 -t 3600 http://localhost:9001/hi
