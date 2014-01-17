package main

import (
	"encoding/json"
	metrics "github.com/rcrowley/go-metrics"
	"net/http"
)

type PingMonitorWeb struct {
	Server   *http.Server
	PingInfo PingInfo
}

func NewPingMonitorWeb(s *http.Server, pi PingInfo) *PingMonitorWeb {
	pm := &PingMonitorWeb{
		Server:   s,
		PingInfo: pi,
	}

	pm.Server.Handler = pm

	return pm
}

func (pm *PingMonitorWeb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	output := map[string]interface{}{}

	for ip, registry := range pm.PingInfo.Registries {
		marshal_tmp := map[string]interface{}{}
		// this seems to be the only way to do this
		// FIXME error checking
		content, err := (*registry).(*metrics.StandardRegistry).MarshalJSON()

		if err != nil {
			w.WriteHeader(500)
			return
		}

		err = json.Unmarshal(content, &marshal_tmp)

		if err != nil {
			w.WriteHeader(500)
			return
		}

		output[ip] = marshal_tmp
	}

	content, _ := json.Marshal(output)
	w.Write(content)
}

func (pm *PingMonitorWeb) Start() error {
	return pm.Server.ListenAndServe()
}
