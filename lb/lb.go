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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	ocp "contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
)

var (
	port     = flag.String("port", "9001", "Port to listen on.")
	backends = flag.String("backends", "", "List of backend addesses, separated by commas, to loadbalance.")
)

var (
	requests   = stats.Int64("requests", "total requests received", stats.UnitDimensionless)
	errors     = stats.Int64("errors", "total errors served", stats.UnitDimensionless)
	latency_ms = stats.Float64(
		"latency_ms",
		"request latency in milliseconds", stats.UnitMilliseconds)
)

var (
	KeyUrl, _   = tag.NewKey("url")
	KeyCode, _  = tag.NewKey("code")
	requestView = &view.View{
		Name:        "requests",
		Measure:     requests,
		Description: "total requests received",
		Aggregation: view.Count()}
	errorsView = &view.View{
		Name:        "errors",
		Measure:     errors,
		Description: "total errors served",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyCode},
	}
	latencyView = &view.View{
		Name:        "latency_ms",
		Measure:     latency_ms,
		Description: "request latency in ms",
		Aggregation: view.Distribution(prometheus.ExponentialBuckets(1, 2, 20)...),
	}
)

var (
	client *http.Client
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	var span *trace.Span
	httpformat := &b3.HTTPFormat{}
	if sc, ok := httpformat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "get", sc)
	} else {
		ctx, span = trace.StartSpan(ctx, "get")
	}
	defer span.End()
	ctx, _ = tag.New(ctx, tag.Insert(KeyUrl, r.URL.Path))
	stats.Record(ctx, requests.M(1)) // COUNTER

	bs := strings.Split(*backends, ",")
	url := fmt.Sprintf("http://%s%s",
		bs[rand.Intn(len(bs))], r.URL.Path)
	span.Annotate([]trace.Attribute{trace.StringAttribute("backend", url)}, "picked backend")
	resp, err := client.Get(url)
	if err != nil {
		stats.RecordWithTags(ctx, []tag.Mutator{
			tag.Upsert(KeyCode, err.Error())}, errors.M(1)) // MAP
		log.Println("get:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		// Copy headers to response
		h := w.Header()
		for k, v := range resp.Header {
			h[k] = v
		}
	} else {
		stats.RecordWithTags(ctx, []tag.Mutator{
			tag.Upsert(KeyCode, http.StatusText(resp.StatusCode))}, errors.M(1)) // MAP
	}
	defer func() {
		l := time.Since(start)
		ms := float64(l.Nanoseconds()) / 1e6
		stats.Record(ctx, latency_ms.M(ms)) // HISTOGRAM
	}()
	w.WriteHeader(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {
	flag.Parse()
	if err := view.Register(requestView, errorsView, latencyView); err != nil {
		log.Fatal(err)
	}
	pe, err := ocp.NewExporter(ocp.Options{})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(pe)
	localEndpoint, err := openzipkin.NewEndpoint("lb", "localhost:"+*port)
	if err != nil {
		log.Fatal(err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411:/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	client = &http.Client{Transport: &ochttp.Transport{}}

	zpages.Handle(http.DefaultServeMux, "/")
	http.HandleFunc("/", handleGet)
	http.Handle("/metrics", pe)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
