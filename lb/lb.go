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

	"github.com/prometheus/client_golang/prometheus"
)

var (
	port     = flag.String("port", "9001", "Port to listen on.")
	backends = flag.String("backends", "", "List of backend addesses, separated by commas, to loadbalance.")
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
)

func init() {
	prometheus.MustRegister(requests)
	prometheus.MustRegister(errors)
	prometheus.MustRegister(latency_ms)
}

var (
	client *http.Client
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requests.Add(1) // COUNTER
	bs := strings.Split(*backends, ",")
	url := fmt.Sprintf("http://%s%s",
		bs[rand.Intn(len(bs))], r.URL.Path)
	resp, err := client.Get(url)
	if err != nil {
		errors.WithLabelValues(err.Error()).Add(1) // MAP
		log.Println("get:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		h := w.Header()
		for k, v := range resp.Header {
			h[k] = v
		}
	} else {
		errors.WithLabelValues(http.StatusText(resp.StatusCode)).Add(1) // MAP
	}
	defer func() {
		l := time.Since(start)
		ms := float64(l.Nanoseconds()) / 1e6
		latency_ms.Observe(ms) // HISTOGRAM
	}()
	w.WriteHeader(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {
	client = &http.Client{}
	flag.Parse()
	http.HandleFunc("/", handleGet)
	http.Handle("/metrics", prometheus.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
