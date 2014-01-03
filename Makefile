all: slides/slides.ps c/c s/s lb/lb
.PHONY: all

slides/%.dvi: slides/%.tex
	latex -output-directory slides $<

slides/%.ps: slides/%.dvi
	dvips -o $@ $<

slides/%.eps: slides/%.svg
	convert $< $@

c/c: c/c.go
	cd c && go build

s/s: s/s.go
	cd s && go build

lb/lb: lb/lb.go
	cd lb && go build
