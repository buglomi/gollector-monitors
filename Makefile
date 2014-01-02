GOPATH="$(shell pwd):$(shell pwd)/gopath"

src/redis-monitor/redis-monitor.go: gopath
	if [ ! -d gopath/src/github.com/vmihailenco/redis ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/vmihailenco/redis; fi
	GOPATH=$(GOPATH) go build redis-monitor

gopath: 
	mkdir -p gopath

.PHONY: gopath
