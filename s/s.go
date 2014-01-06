package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var (
	port = flag.String("port", "8000", "Port to listen on.")
)

var (
	requests = expvar.NewInt("requests")
	errors   = expvar.NewInt("errors")
	// Buckets of log2 lower bound latency.
	latency    = expvar.NewMap("latency")
	latency_ms = expvar.NewMap("latency_ms")
)

var (
	zipf = rand.NewZipf(rand.New(rand.NewSource(0)), 1.1, 1, 1000)
)

func handleHi(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requests.Add(1)

	// Perform a "database" "lookup".
	time.Sleep(time.Duration(zipf.Uint64()) * time.Millisecond)

	// Fail sometimes.
	if rand.Intn(100) > 95 {
		w.WriteHeader(500)
		return
	}

	// Record metrics.
	defer func() {
		l := time.Since(start)
		bucket := fmt.Sprintf("%.0f", math.Exp2(math.Logb(float64(l.Nanoseconds()/1e6))))
		latency.Add(bucket, 1)
		latency_ms.Add(bucket, l.Nanoseconds()/1e6)
	}()

	// Return page content.
	w.Write([]byte("hi\n"))
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/hi", handleHi)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
