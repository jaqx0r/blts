all: slides.ps c/c s/s lb/lb
.PHONY: all

%.dvi: %.tex
	latex $<

%.ps: %.dvi
	dvips $<

%.eps: %.svg
	convert $< $@

c/c: c/c.go
	cd c && go build

s/s: s/s.go
	cd s && go build

lb/lb: lb/lb.go
	cd lb && go build
