package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

var (
	port     = flag.String("port", "8080", "Port to listen on.")
	backends = flag.String("backends", "", "List of backend addesses, separated by commas, to loadbalance.")
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	bs := strings.Split(*backends, ",")
	url := fmt.Sprintf("http://%s/%s", bs[rand.Intn(len(bs))], r.URL.Path)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	h := w.Header()
	for k, v := range resp.Header {
		h[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handleGet)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
