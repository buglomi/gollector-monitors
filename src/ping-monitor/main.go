package main

import (
	"flag"
	"github.com/gollector/gollector-monitors/src/httpmetrics"
	metrics "github.com/rcrowley/go-metrics"
	"os"
)

func main() {
	count := flag.Int("count", 5, "The number of pings to send")
	wait := flag.Float64("wait", 15, "The amount of time to wait for all pings to come back in seconds")
	interval := flag.Float64("interval", 2, "The interval to wait between pings in seconds")
	repeat := flag.Int("repeat", 30, "Repeat the whole process every x seconds")
	socket := flag.String("socket", "/tmp/ping-monitor.sock", "Unix socket to create/listen on")

	flag.Parse()

	if len(flag.Args()) < 1 {
		os.Stderr.WriteString("Must supply at least one IP address!\n")
		os.Exit(1)
	}

	h := &httpmetrics.Handler{
		Socket:     *socket,
		Registries: make(map[string]*metrics.Registry),
	}

	pi := &PingInfo{
		Count:      *count,
		Wait:       *wait,
		Interval:   *interval,
		Repeat:     *repeat,
		Hosts:      flag.Args(),
		Registries: &(h.Registries),
	}

	InitPing(pi)

	if err := h.CreateServer(); err != nil {
		panic(err)
	}
}
