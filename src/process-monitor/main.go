package main

import (
	"encoding/json"
	"fmt"
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
	if len(os.Args) < 2 {
		fmt.Println("enter the full path to a binary")
		os.Exit(1)
	}

	s := &http.Server{
		Addr: "127.0.0.1:9118",
		Handler: &PMHandler{
			Binaries: os.Args[1:],
		},
	}

	err := s.ListenAndServe()

	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
