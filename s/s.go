package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	port = flag.String("port", "8080", "Port to listen on.")
)

func main() {
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
