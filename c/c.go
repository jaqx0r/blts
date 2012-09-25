package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	socketpath = flag.String("socketpath", "/var/run/collectd-unixsock", "Path to collectd unixsock.")
	targets    = flag.String("targets", "", "comma separated list of hostports to collect.")
)

const (
	INTERVAL        = 1 // Seconds
	COLLECTD_FORMAT = "PUTVAL \"gunstar/demo-%s/%s-%s\" interval=%d %d:%d\n"
)

func write(lines chan string) {
	collectd, err := net.Dial("unix", *socketpath)
	if err != nil {
		log.Fatal(err)
	}
	defer collectd.Close()
	for {
		select {
		case line := <-lines:
			_, err := fmt.Fprint(collectd, line)
			if err != nil {
				ioutil.ReadAll(collectd)
			}
		}
	}
}

func fetch(t string, lines chan string) {
	resp, err := http.Get(fmt.Sprintf("http://%s/debug/vars", t))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var j map[string]interface{}
	err = d.Decode(&j)
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range j {
		switch k {
		case "latency", "requests", "errors":
		// Carry on.
		default:
			continue
		}
		switch n := v.(type) {
		case float64:
			s := fmt.Sprintf(COLLECTD_FORMAT,
				t,
				"counter",
				k,
				INTERVAL,
				time.Now().Unix(),
				int64(n))
			lines <- s
		case map[string]interface{}:
			for k1, v1 := range n {
				switch n1 := v1.(type) {
				case float64:
					s := fmt.Sprintf(COLLECTD_FORMAT,
						t,
						"counter",
						fmt.Sprintf("%s_%s", k, k1),
						INTERVAL,
						time.Now().Unix(),
						int64(n1))
					lines <- s
				}
			}
		}
	}
}

func main() {
	flag.Parse()

	ts := strings.Split(*targets, ",")

	lines := make(chan string, 1)

	go write(lines)

	for {
		select {
		case <-time.After(INTERVAL * time.Second):
			for _, t := range ts {
				go fetch(t, lines)
			}
		}
	}
}
