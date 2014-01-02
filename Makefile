GOPATH="$(shell pwd):$(shell pwd)/gopath"

redis-monitor: gopath
	if [ ! -d gopath/src/github.com/vmihailenco/redis ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/vmihailenco/redis; fi
	GOPATH=$(GOPATH) go build redis-monitor

gopath: 
	mkdir -p gopath

.PHONY: gopath
