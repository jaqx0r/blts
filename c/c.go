package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	targets = flag.String("targets", "", "comma separated list of hostports to collect.")
)

const (
	INTERVAL        = 1                // Seconds
	FILENAME_FORMAT = "data/%s-%s.csv" // target, variable
)

type record struct {
	target string
	name   string
	values []string
}

var cs map[string]*csv.Writer

func write(records chan record) {
	for {
		select {
		case r := <-records:
			filename := fmt.Sprintf(FILENAME_FORMAT, r.target, r.name)
			if c, ok := cs[filename]; ok {
				fmt.Printf("writing values to %s\n", filename)
				c.Write(r.values)
				c.Flush()
			} else {
				var err error
				var f *os.File
				f, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
				if err != nil {
					log.Println("Couldn't open %q, %s", filename, err)
					break
				}
				fmt.Printf("opened %s\n", filename)
				cs[filename] = csv.NewWriter(f)
				cs[filename].Write(r.values)
				cs[filename].Flush()
			}
		}
	}
}

func fetch(t string, lines chan record) {
	resp, err := http.Get(fmt.Sprintf("http://%s/debug/vars", t))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var j map[string]interface{}
	err = d.Decode(&j)
	if err != nil {
		fmt.Println(err)
		return
	}
	now := fmt.Sprintf("%d", time.Now().Unix())
	for k, v := range j {
		switch k {
		case "latency":
			vals := v.(map[string]interface{})
			r := []string{now}
			// Pad out the map with missing values
			buckets := []int64{0, 1, 2, 4, 6, 8, 16, 32, 64, 128, 256, 512}
			for _, b := range buckets {
				if v1, ok := vals[fmt.Sprintf("%d", b)]; ok {
					r = append(r, fmt.Sprintf("%v", v1))
				} else {
					r = append(r, "0")
				}
			}
			lines <- record{t, k, r}
		case "requests", "errors":
			r := []string{now, fmt.Sprintf("%v", v)}
			lines <- record{t, k, r}
		default:
			// Ignore
		}
	}
}

func main() {
	flag.Parse()
	cs = make(map[string]*csv.Writer, 0)

	ts := strings.Split(*targets, ",")

	records := make(chan record, 100)

	go write(records)

	for {
		select {
		case <-time.After(INTERVAL * time.Second):
			for _, t := range ts {
				go fetch(t, records)
			}
			for _, c := range cs {
				c.Flush()
			}
		}
	}
}
