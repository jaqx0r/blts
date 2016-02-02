all: s/s lb/lb
.PHONY: all

s/s: s/s.go
	cd s && go build

lb/lb: lb/lb.go
	cd lb && go build

rules := $(wildcard prom/*.rules)

check-rules: $(rules)
	~/go/src/github.com/prometheus/prometheus/promtool check-rules $(rules)

check-config: prom/prometheus.yml
	~/go/src/github.com/prometheus/prometheus/promtool check-config $<
