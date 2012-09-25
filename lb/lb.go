package main

import (
	"expvar"
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

var (
	requests = expvar.NewInt("requests")
	errors   = expvar.NewInt("errors")
)

var (
	client *http.Client
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	requests.Add(1)
	bs := strings.Split(*backends, ",")
	url := fmt.Sprintf("http://%s%s",
		bs[rand.Intn(len(bs))], r.URL.Path)
	resp, err := client.Get(url)
	if err != nil {
		errors.Add(1)
		log.Println(err)
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
		errors.Add(1)
	}
	w.WriteHeader(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {
	client = &http.Client{}
	flag.Parse()
	http.HandleFunc("/", handleGet)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
