#!/bin/sh
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

trap "killall lb; killall s" HUP TERM INT EXIT

export GOMAXPROCS=4

renice +10 $$

BACKENDS=""
for i in $(seq 0 9); do
    # fire up some old-school containers
    (s/s --port 800$i $@ 2>&1 | sed -e "s/^/s:800$i: /") &
    BACKENDS="$BACKENDS :800$i"
done

BACKENDS=$(echo $BACKENDS | tr ' ' ',')
(lb/lb --backends $BACKENDS $@ 2>&1 | sed -e "s/^/lb: /") &

wait
