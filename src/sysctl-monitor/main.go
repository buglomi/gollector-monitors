package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"util"
)

var registry = map[string]float64{}
var mutex = new(sync.RWMutex)

func Handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer mutex.RUnlock()
	mutex.RLock()
	content, err := json.Marshal(registry)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write(content)
}

func populateRegistry() {
	for {
		mutex.Lock()
		for key, _ := range registry {
			newkey := strings.Replace(key, ".", "/", -1)

			// /proc/sys is what the manpage recommends over sysctl(2)
			file, err := os.Open(filepath.Join("/proc/sys", newkey))

			if err != nil {
				panic(err)
			}

			content, err := ioutil.ReadAll(file)
			file.Close()

			if err != nil {
				fmt.Println(err)
				registry[key] = 0
			}

			result, err := strconv.ParseFloat(strings.Trim(string(content), "\n"), 64)

			if err != nil {
				fmt.Println(err)
				registry[key] = 0
			}

			registry[key] = result
		}
		mutex.Unlock()

		time.Sleep(60 * time.Second)
	}
}

func main() {
	socket := flag.String("socket", "/tmp/sysctl-monitor.sock", "UNIX Socket to expose")
	flag.Parse()

	for _, val := range flag.Args() {
		mutex.Lock()
		registry[val] = 0
		mutex.Unlock()
	}

	go populateRegistry()

	http.HandleFunc("/", Handler)

	l, err := util.CreateSocket(*socket)

	if err != nil {
		panic(err)
	}

	panic(http.Serve(l, nil))
}
