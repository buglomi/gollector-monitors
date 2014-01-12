package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type PingInfo struct {
	Count    int
	Wait     float64
	Interval float64
}

func main() {
	count := flag.Int("count", 10, "The number of pings to send")
	wait := flag.Float64("wait", 10, "The amount of time to wait for all pings to come back in seconds")
	interval := flag.Float64("interval", 1, "The interval to wait between pings in seconds")

	flag.Parse()

	if len(flag.Args()) < 1 {
		os.Stderr.WriteString("Must supply at least one IP address!\n")
		os.Exit(1)
	}

	pi := PingInfo{
		Count:    *count,
		Wait:     *wait,
		Interval: *interval,
	}

	num := len(flag.Args())
	done := make(chan bool, 100)

	for _, ip := range flag.Args() {
		// ping each host listed -- print when complete
		go func(ip string) {
			result := pi.connectAndPing(ip)
			fmt.Println(ip+":", len(result), result)
			done <- true
		}(ip)
	}

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
