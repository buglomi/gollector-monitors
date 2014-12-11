package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gollector/gollector-monitors/src/util"
	"net/http"
	"os"
)

type PMHandler struct {
	Binaries []string
}

func (p *PMHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	content, err := json.Marshal(GetPids(p.Binaries...))

	if err != nil {
		w.Write([]byte("null"))
		return
	}

	w.Write(content)
}

func main() {
	socket := flag.String("socket", "/tmp/process-monitor.sock", "Path to the socket we serve metrics over")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("enter the full path to a binary")
		os.Exit(1)
	}

	s := &http.Server{
		Handler: &PMHandler{
			Binaries: flag.Args(),
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
