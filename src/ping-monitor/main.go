package main

import (
	"flag"
	metrics "github.com/rcrowley/go-metrics"
	"log"
	"os"
	"time"
)

type PingInfo struct {
	Count    int
	Wait     float64
	Interval float64
}

func main() {
	count := flag.Int("count", 100, "The number of pings to send")
	wait := flag.Float64("wait", 20, "The amount of time to wait for all pings to come back in seconds")
	interval := flag.Float64("interval", 1, "The interval to wait between pings in seconds")

	flag.Parse()

	if len(flag.Args()) < 1 {
		os.Stderr.WriteString("Must supply at least one IP address!\n")
		os.Exit(1)
	}

	registries := map[string]metrics.Registry{}

	pi := PingInfo{
		Count:    *count,
		Wait:     *wait,
		Interval: *interval,
	}

	num := len(flag.Args())
	done := make(chan bool, 100)

	for _, ip := range flag.Args() {
		registries[ip] = metrics.NewRegistry()
		registry := registries[ip]
		go metrics.Log(registry, 1*time.Second, log.New(os.Stderr, ip+": ", log.Lmicroseconds))

		for i := 0; i < 10; i++ {
			// ping each host listed -- print when complete
			go func(ip string, registry *metrics.Registry) {
				pi.connectAndPing(ip, registry)
				done <- true
			}(ip, &registry)
			time.Sleep(10 * time.Second)
		}
	}

	num = num * 10

	// block until we're done
	for num > 0 {
		select {
		case <-done:
			num--
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}
