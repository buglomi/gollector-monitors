package main

import (
	"flag"
	"net/http"
	"os"
)

func main() {
	count := flag.Int("count", 10, "The number of pings to send")
	wait := flag.Float64("wait", 20, "The amount of time to wait for all pings to come back in seconds")
	interval := flag.Float64("interval", 1, "The interval to wait between pings in seconds")
	repeat := flag.Int("repeat", 60, "Repeat the whole process every x seconds")

	flag.Parse()

	if len(flag.Args()) < 1 {
		os.Stderr.WriteString("Must supply at least one IP address!\n")
		os.Exit(1)
	}

	pi := PingInfo{
		Count:    *count,
		Wait:     *wait,
		Interval: *interval,
		Repeat:   *repeat,
		Hosts:    flag.Args(),
	}

	InitPing(pi)

	s := &http.Server{
		Addr: "127.0.0.1:9119",
	}

	pmw := NewPingMonitorWeb(s)

	err := pmw.Start()

	if err != nil {
		panic(err)
		os.Exit(1)
	}
}
