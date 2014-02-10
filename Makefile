PACKAGES=\
				 github.com/vmihailenco/redis\
				 github.com/bmizerany/pq\
				 github.com/rcrowley/go-metrics


MONITORS=\
				 redis-monitor\
				 postgresql-monitor\
				 ping-monitor\
				 process-monitor\
				 tcp-monitor\
				 sysctl-monitor

GOPATH="$(shell pwd):$(shell pwd)/gopath"

all: $(MONITORS)

dist: all
	tar czf gollector-monitors.tar.gz $(MONITORS)

redis-monitor: goget
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build redis-monitor

postgresql-monitor: goget
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build postgresql-monitor

ping-monitor: goget 
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build ping-monitor

process-monitor: goget
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build process-monitor

tcp-monitor: goget
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build tcp-monitor

sysctl-monitor: goget
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build sysctl-monitor

goget: godep gopath
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:$(shell pwd)" godep get $(PACKAGES)

godepsave: 
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:$(shell pwd)" godep save $(PACKAGES)

godep: gopath
	if [ ! -x gopath/bin/godep ]; then /usr/bin/env GOPATH=gopath go get -u -d github.com/kr/godep; fi
	GOPATH=$(shell pwd):$(shell pwd)/gopath go install github.com/kr/godep

gopath: 
	mkdir -p gopath

.PHONY: gopath
