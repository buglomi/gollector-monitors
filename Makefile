GOPATH="$(shell pwd):$(shell pwd)/gopath"

all: redis-monitor postgresql-monitor ping-monitor

redis-monitor: gopath
	if [ ! -d gopath/src/github.com/vmihailenco/redis ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/vmihailenco/redis; fi
	GOPATH=$(GOPATH) go build redis-monitor

postgresql-monitor: gopath
	if [ ! -d gopath/src/github.com/bmizerany/pq ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/bmizerany/pq; fi
	GOPATH=$(GOPATH) go build postgresql-monitor

ping-monitor: gopath
	GOPATH=$(GOPATH) go build ping-monitor

gopath: 
	mkdir -p gopath

.PHONY: gopath
