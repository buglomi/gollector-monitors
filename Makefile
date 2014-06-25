PACKAGES=\
				 github.com/vmihailenco/redis\
				 github.com/bmizerany/pq\
				 github.com/rcrowley/go-metrics\
				 github.com/bitly/nsq/util/lookupd\


MONITORS=\
				 redis-monitor\
				 postgresql-monitor\
				 ping-monitor\
				 process-monitor\
				 tcp-monitor\
				 sysctl-monitor\
				 http-monitor\
				 nsq-monitor

all: ${MONITORS}

clean:
	rm -rf gopath
	rm -rf Godeps/_workspace
	rm -f ${MONITORS}

rebuild: clean all

godepsave: gopath/bin/godep
	PATH="$(PATH):gopath/bin" GOPATH="${PWD}/gopath:${PWD}" godep save ${PACKAGES}

dist: clean all
	tar czf gollector-monitors.tar.gz ${MONITORS}

${MONITORS}: %: gopath/src/built
	PATH="${PATH}:gopath/bin" GOPATH="${PWD}/gopath:Godeps/_workspace:${PWD}" godep go build $*

gopath/src/built: gopath gopath/bin/godep
	PATH="${PATH}:gopath/bin" GOPATH="${PWD}/gopath:${PWD}" godep get ${PACKAGES}
	touch gopath/src/built

gopath/bin/godep: gopath
	/usr/bin/env GOPATH=gopath go get -u -d github.com/kr/godep
	GOPATH=${PWD}:${PWD}/gopath go install github.com/kr/godep

gopath:
	mkdir -p gopath

.PHONY: godepsave
