package main

import (
	"fmt"
	metrics "github.com/rcrowley/go-metrics"
	"httpmetrics"
	"net"
	"os"
	"time"
)

type Addr struct {
	Address  string
	Registry *metrics.Registry
}

var addrs = []*Addr{}

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide one or more IP:Port pairs")
		os.Exit(1)
	}

	h := &httpmetrics.Handler{
		Registries: make(map[string]*metrics.Registry),
		Socket:     "/tmp/tcp-monitor.sock",
	}

	for _, addr := range os.Args[1:] {
		r := metrics.NewRegistry()
		a := &Addr{
			Address:  addr,
			Registry: &r,
		}

		addrs = append(addrs, a)
		a.startPing()
		h.Registries[addr] = &r
	}

	if err := h.CreateServer(); err != nil {
		panic(err)
	}
}
