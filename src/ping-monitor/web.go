package main

import (
	"net/http"
)

type PingMonitorWeb struct {
	Server   *http.Server
	PingInfo PingInfo
}

func NewPingMonitorWeb(s *http.Server) *PingMonitorWeb {
	return &PingMonitorWeb{
		Server: s,
	}
}

func (pm *PingMonitorWeb) Start() error {
	return pm.Server.ListenAndServe()
}
