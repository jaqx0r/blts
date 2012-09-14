package main

import (
	"bufio"
	_ "expvar"
	"flag"
	"log"
	"net/http"
)

var (
	port = flag.String("port", "8080", "Port to listen on.")
)

func handleHi(w http.ResponseWriter, r *http.Request) {
	b := bufio.NewWriter(w)
	defer b.Flush()
	b.WriteString("hi")
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/", handleHi)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
