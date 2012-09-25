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
			// fmt.Print(line)
			_, err := fmt.Fprint(collectd, line)
			if err != nil {
				ioutil.ReadAll(collectd)
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
				resp, err := http.Get(fmt.Sprintf("http://%s/debug/vars", t))
				if err != nil {
					log.Println(err)
				}
				d := json.NewDecoder(resp.Body)
				var j map[string]interface{}
				err := d.Decode(&j)
				for k, v := range j {
					switch v.(type) {
					case float64, int64:
						s := fmt.Sprintf(COLLECTD_FORMAT,
							t,
							"counter",
							k,
							INTERVAL,
							time.Now().Unix(),
							v)
						lines <- s
case int
					}
				}
			}
		}
	}
}
