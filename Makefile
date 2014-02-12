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

clean:
	rm -rf gopath
	rm -rf Godeps/_workspace
	rm -f $(MONITORS)

rebuild: clean all

godepsave: gopath/bin/godep Godeps/_workspace/src
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:$(shell pwd)" godep save $(PACKAGES)

dist: clean all
	tar czf gollector-monitors.tar.gz $(MONITORS)

${MONITORS}: %: gopath/src/built
	PATH="${PATH}:gopath/bin" GOPATH="$(shell pwd)/gopath:Godeps/_workspace:$(shell pwd)" godep go build $*

gopath/src/built: gopath gopath/bin/godep
	PATH="$(PATH):gopath/bin" GOPATH="$(shell pwd)/gopath:$(shell pwd)" godep get $(PACKAGES)
	touch gopath/src/built

gopath/bin/godep: gopath
	/usr/bin/env GOPATH=gopath go get -u -d github.com/kr/godep
	GOPATH=$(shell pwd):$(shell pwd)/gopath go install github.com/kr/godep

gopath:
	mkdir -p gopath

.PHONY: godepsave
