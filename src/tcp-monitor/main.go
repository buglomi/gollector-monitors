package main

import (
	"encoding/json"
	"fmt"
	metrics "github.com/rcrowley/go-metrics"
	"net"
	"net/http"
	"os"
	"time"
)

var addrs = []*Addr{}

type Addr struct {
	Address  string
	Registry *metrics.Registry
}

var registries = map[string]*metrics.Registry{}

func (a *Addr) startPing() {
	go func(a *Addr) {
		for {
			a.ping()
			time.Sleep(1 * time.Second)
		}
	}(a)
}

func (a *Addr) ping() {
	errors := int64(0)
	start := time.Now()
	conn, err := net.Dial("tcp", a.Address)
	if err != nil {
		errors = 100
	} else {
		conn.Close()
		metrics.GetOrRegisterHistogram(
			"ns",
			*a.Registry,
			metrics.NewUniformSample(60),
		).Update(time.Since(start).Nanoseconds())
	}

	metrics.GetOrRegisterHistogram(
		"errors",
		*a.Registry,
		metrics.NewUniformSample(60),
	).Update(errors)
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	output := map[string]interface{}{}

	for _, addr := range addrs {
		marshal_tmp := map[string]interface{}{}
		content, err := (*addr.Registry).(*metrics.StandardRegistry).MarshalJSON()

		if err != nil {
			w.WriteHeader(500)
			return
		}

		err = json.Unmarshal(content, &marshal_tmp)

		if err != nil {
			w.WriteHeader(500)
			return
		}

		output[addr.Address] = marshal_tmp
	}

	content, _ := json.Marshal(output)
	w.Write(content)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide one or more IP:Port pairs")
		os.Exit(1)
	}

	for _, addr := range os.Args[1:] {
		r := metrics.NewRegistry()
		a := &Addr{
			Address:  addr,
			Registry: &r,
		}

		addrs = append(addrs, a)
		a.startPing()
	}

	http.HandleFunc("/", handlerFunc)

	s := http.Server{}

	c, err := net.Dial("unix", "/tmp/tcp-monitor.sock")

	if err == nil {
		c.Close()
		panic("socket in use")
	} else {
		os.Remove("/tmp/tcp-monitor.sock")
	}

	l, err := net.Listen("unix", "/tmp/tcp-monitor.sock")

	if err != nil {
		panic(err)
	}

	err = s.Serve(l)

	if err != nil {
		panic(err)
	}
}
