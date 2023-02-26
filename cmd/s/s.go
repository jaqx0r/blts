// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	ocp "contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
)

var (
	port       = flag.String("port", "8000", "Port to listen on.")
	faily      = flag.Bool("faily", false, "Fail more often.")
	zipkinAddr = flag.String("zipkin", "localhost:9411", "Zipkin address")
)

var (
	requests           = stats.Int64("requests", "total requests received", stats.UnitDimensionless)
	errors             = stats.Int64("errors", "total errors served", stats.UnitDimensionless)
	latency_ms         = stats.Float64("latency_ms", "request latency in milliseconds", stats.UnitMilliseconds)
	backend_latency_ms = stats.Float64("backend_latency_ms", "backend request latency in milliseconds", stats.UnitMilliseconds)
)

var (
	KeyUrl, _  = tag.NewKey("url")
	KeyCode, _ = tag.NewKey("code")

	requestView = &view.View{Name: "requests",
		Measure:     requests,
		Description: "total requests received eh",
		Aggregation: view.Count(),
	}
	errorsView = &view.View{Name: "errors",
		Measure:     errors,
		Description: "total errors servved",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyCode},
	}
	latencyView = &view.View{
		Name:        "latency_ms",
		Measure:     latency_ms,
		Description: "request latency in ms",
		Aggregation: view.Distribution(prometheus.ExponentialBuckets(1, 2, 20)...),
	}
	backendLatencyView = &view.View{
		Name:        "backend_latency_ms",
		Measure:     backend_latency_ms,
		Description: "backend request latency in ms",
		Aggregation: view.Distribution(prometheus.ExponentialBuckets(1, 2, 20)...),
	}
)

var (
	randLock sync.Mutex
	zipf     = rand.NewZipf(rand.New(rand.NewSource(0)), 1.1, 1, 1000)
)

// Perform a "database" "lookup".
func databaseCall(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "databaseCall")
	defer span.End()

	backend_start := time.Now()
	defer func() {
		stats.Record(ctx,
			backend_latency_ms.M((float64(time.Since(backend_start).Nanoseconds() / 1e6)))) // HISTOGRAM
	}()

	t := time.Duration(zipf.Uint64()) * time.Millisecond
	randLock.Lock() // golang issue 3611
	time.Sleep(t)
	randLock.Unlock()
	span.Annotatef(nil, "databased for %s", t)
}

func handleHi(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	var span *trace.Span
	httpformat := &b3.HTTPFormat{}
	if sc, ok := httpformat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "hi", sc)
	} else {
		ctx, span = trace.StartSpan(ctx, "hi")
	}
	defer span.End()
	ctx, _ = tag.New(ctx, tag.Insert(KeyUrl, "hi"))
	stats.Record(ctx, requests.M(1)) // COUNTER

	span.Annotate(nil, "gunna database")
	databaseCall(ctx)
	span.Annotate(nil, "databased")

	// Fail sometimes.
	switch v := rand.Intn(100); {
	case v >= 99:
		span.Annotate(nil, "v > 99")
		stats.RecordWithTags(ctx, []tag.Mutator{
			tag.Upsert(KeyCode, http.StatusText(500))}, errors.M(1)) // MAP
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: "bad dice roll"})
		w.WriteHeader(500)
		return
	case v >= 90:
		span.Annotate(nil, "v > 90")
		if *faily {
			span.Annotate(nil, "extra fail flag set")
			stats.RecordWithTags(ctx, []tag.Mutator{
				tag.Upsert(KeyCode, http.StatusText(500))}, errors.M(1)) // MAP
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: "extra faily"})
			w.WriteHeader(500)
			return
		}
	}
	span.Annotate(nil, "v  ok")
	span.SetStatus(trace.Status{Code: trace.StatusCodeOK, Message: "OK"})

	// Record metrics.
	defer func() {
		l := time.Since(start)
		ms := float64(l.Nanoseconds()) / 1e6
		stats.Record(ctx, latency_ms.M(ms)) // HISTOGRAM
	}()

	// Return page content.
	w.Write([]byte("hi\n"))
}

func main() {
	flag.Parse()

	if err := view.Register(requestView, errorsView, latencyView, backendLatencyView); err != nil {
		log.Fatal(err)
	}
	pe, err := ocp.NewExporter(ocp.Options{})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(pe)
	if *zipkinAddr != "" {
		localEndpoint, err := openzipkin.NewEndpoint("s", "localhost:"+*port)
		if err != nil {
			log.Fatal(err)
		}
		reporter := zipkinHTTP.NewReporter(fmt.Sprintf("http://%s/api/v2/spans", *zipkinAddr))
		ze := zipkin.NewExporter(reporter, localEndpoint)
		trace.RegisterExporter(ze)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/hi", handleHi)
	http.Handle("/metrics", pe)
	zpages.Handle(http.DefaultServeMux, "/")
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
