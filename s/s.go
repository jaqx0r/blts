package main

import (
	"bufio"
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
	port = flag.String("port", "8080", "Port to listen on.")
)

var (
	requests = expvar.NewInt("requests")
	errors   = expvar.NewInt("errors")
	// Buckets of lower bound latency in log2 buckets.
	latency = expvar.NewMap("latency")
)

var (
	zipf = rand.NewZipf(rand.New(rand.NewSource(0)), 1.1, 1, 1000)
)

func handleHi(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requests.Add(1)
	// Perform a database lookup.
	time.Sleep(time.Duration(zipf.Uint64()) * time.Millisecond)
	if rand.Intn(100) > 99 {
		w.WriteHeader(500)
		return
	}
	defer func() {
		l := time.Since(start)
		latency.Add(fmt.Sprintf("%.0f", math.Exp2(math.Logb(float64(l.Nanoseconds()/1e6)))), 1)
	}()
	b := bufio.NewWriter(w)
	defer b.Flush()
	b.WriteString("hi")
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/hi", handleHi)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
