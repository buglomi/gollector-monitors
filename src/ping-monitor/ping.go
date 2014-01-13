package main

import (
	metrics "github.com/rcrowley/go-metrics"
	"log"
	"os"
	"time"
)

func InitPing(pi PingInfo) {
	registries := map[string]metrics.Registry{}

	for _, ip := range pi.Hosts {
		registries[ip] = metrics.NewRegistry()
		registry := registries[ip]
		go metrics.Log(registry, time.Duration(pi.Repeat)*time.Second, log.New(os.Stderr, ip+": ", log.Lmicroseconds))

		// ping each host listed -- print when complete
		go func(ip string, registry *metrics.Registry) {
			for {
				pi.connectAndPing(ip, registry)
				time.Sleep(time.Duration(pi.Repeat) * time.Second)
			}
		}(ip, &registry)
	}
}
