package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bitly/nsq/util/lookupd"
	"net/http"
	"os"
	"util"
)

type NSQHandler struct {
	NSQLookupdHTTPAddresses []string
}

type ChannelData struct {
	Depth         int64
	MemoryDepth   int64
	BackendDepth  int64
	InFlightCount int64
	DeferredCount int64
	RequeueCount  int64
	TimeoutCount  int64
	MessageCount  int64
	ClientCount   int
}

func getTopicStats(topicName string, NSQLookupdHTTPAddresses []string) map[string]ChannelData {
	var producers []string
	producers, _ = lookupd.GetLookupdTopicProducers(topicName, NSQLookupdHTTPAddresses)
	_, channelStats, _ := lookupd.GetNSQDStats(producers, topicName)
	channelMap := make(map[string]ChannelData)
	for _, c := range channelStats {
		channelMap[c.ChannelName] = ChannelData{
			Depth:         c.Depth,
			MemoryDepth:   c.MemoryDepth,
			BackendDepth:  c.BackendDepth,
			InFlightCount: c.InFlightCount,
			DeferredCount: c.DeferredCount,
			RequeueCount:  c.RequeueCount,
			TimeoutCount:  c.TimeoutCount,
			MessageCount:  c.MessageCount,
			ClientCount:   c.ClientCount,
		}
	}
	return channelMap
}

func (n *NSQHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var topics []string
	topicStats := make(map[string]interface{})
	topics, _ = lookupd.GetLookupdTopics(n.NSQLookupdHTTPAddresses)
	for _, topic := range topics {
		topicStats[topic] = getTopicStats(topic, n.NSQLookupdHTTPAddresses)
	}
	content, _ := json.Marshal(topicStats)
	w.Write(content)
}

func main() {
	socket := flag.String("socket", "/tmp/nsq-monitor.sock", "Path to the socket we serve metrics over")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Enter one or more NSQ Lookupd addresses (ex: 10.0.0.1:4161)")
		os.Exit(1)
	}

	s := &http.Server{
		Handler: &NSQHandler{
			NSQLookupdHTTPAddresses: flag.Args(),
		},
	}

	l, err := util.CreateSocket(*socket)

	if err != nil {
		panic(err)
	}

	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
