package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"util"
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
	if len(os.Args) < 2 {
		fmt.Println("enter the full path to a binary")
		os.Exit(1)
	}

	s := &http.Server{
		Handler: &PMHandler{
			Binaries: os.Args[1:],
		},
	}

	l, err := util.CreateSocket("/tmp/process-monitor.sock")

	if err != nil {
		panic(err)
	}

	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
