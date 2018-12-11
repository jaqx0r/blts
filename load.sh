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

renice +10 $$

max=10
for i in $(seq 1 $(( 2 * $max )) ); do
    (ab -r -n 100000 -c 1000 http://localhost:9001/hi 2>&1 | sed -e "s/^/$i: /") &
    sleep $(echo $(( $max - $i )) | tr -d '-')
done

wait
