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

all: ./s ./lb
.PHONY: all

clean:
	rm -rf ./s ./lb
.PHONY: clean

./s: cmd/s/s.go
	go build ./cmd/s

./lb: cmd/lb/lb.go
	go build ./cmd/lb

rules := $(wildcard prom/*.rules)

check-rules: $(rules)
	~/go/src/github.com/prometheus/prometheus/promtool check-rules $(rules)

check-config: prom/prometheus.yml
	~/go/src/github.com/prometheus/prometheus/promtool check-config $<
