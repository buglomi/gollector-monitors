MONITORS=\
				 redis-monitor\
				 postgresql-monitor\
				 ping-monitor\
				 process-monitor\
				 tcp-monitor

GOPATH="$(shell pwd):$(shell pwd)/gopath"

all: $(MONITORS)

dist: all
	tar czf gollector-monitors.tar.gz $(MONITORS)

redis-monitor: gopath
	if [ ! -d gopath/src/github.com/vmihailenco/redis ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/vmihailenco/redis; fi
	GOPATH=$(GOPATH) go build redis-monitor

postgresql-monitor: gopath
	if [ ! -d gopath/src/github.com/bmizerany/pq ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/bmizerany/pq; fi
	GOPATH=$(GOPATH) go build postgresql-monitor

ping-monitor: gopath
	if [ ! -d gopath/github.com/rcrowley/go-metrics ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/rcrowley/go-metrics; fi
	GOPATH=$(GOPATH) go build ping-monitor

process-monitor: gopath
	GOPATH=$(GOPATH) go build process-monitor

tcp-monitor: gopath
	if [ ! -d gopath/github.com/rcrowley/go-metrics ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/rcrowley/go-metrics; fi
	GOPATH=$(GOPATH) go build tcp-monitor

gopath: 
	mkdir -p gopath

.PHONY: gopath
