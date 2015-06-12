all: s/s lb/lb
.PHONY: all

s/s: s/s.go
	cd s && go build

lb/lb: lb/lb.go
	cd lb && go build
