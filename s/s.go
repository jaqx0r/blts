package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	port = flag.String("port", "8000", "Port to listen on.")
)

var (
	requests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests", Help: "total requests received"})
	errors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "errors", Help: "total errors served"}, []string{"code"})
	latency_ms = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "latency_ms",
		Help:    "request latency in milliseconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20)})
	backend_latency_ms = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "backend_latency_ms",
		Help:    "request latency in milliseconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20)})
)

func init() {
	prometheus.MustRegister(requests)
	prometheus.MustRegister(errors)
	prometheus.MustRegister(latency_ms)
	prometheus.MustRegister(backend_latency_ms)
}

var (
	//randLock sync.Mutex
	zipf = rand.NewZipf(rand.New(rand.NewSource(0)), 1.1, 1, 1000)
)

func handleHi(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requests.Add(1) // COUNTER

	// Perform a "database" "lookup".
	backend_start := time.Now()
	//randLock.Lock() // golang issue 3611
	time.Sleep(time.Duration(zipf.Uint64()) * time.Millisecond)
	//randLock.Unlock()
	backend_latency_ms.Observe(float64(time.Since(backend_start).Nanoseconds() / 1e6)) // HISTOGRAM

	// Fail sometimes.
	switch v := rand.Intn(100); {
	case v > 95:
		errors.WithLabelValues(http.StatusText(500)).Add(1) // MAP
		w.WriteHeader(500)
		return
	case v > 85:
		errors.WithLabelValues(http.StatusText(400)).Add(1) // MAP
		w.WriteHeader(400)
		return
	}

	// Record metrics.
	defer func() {
		l := time.Since(start)
		ms := float64(l.Nanoseconds()) / 1e6
		latency_ms.Observe(ms) // HISTOGRAM
	}()

	// Return page content.
	w.Write([]byte("hi\n"))
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/hi", handleHi)
	http.Handle("/metrics", prometheus.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
