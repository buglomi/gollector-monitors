package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davecheney/profile"
	"github.com/fzzy/radix/redis"
	"github.com/gollector/gollector-monitors/src/util"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)
import _ "net/http/pprof"

type DBKeyspace struct {
	Keys    int64
	Expires int64
}

type RedisStats struct {
	FreeMemory             int64
	EvictedKeys            int64
	InstantaneousOpsPerSec int64
	ClientConnections      int64
	DBs                    map[string]*DBKeyspace
}

type Addresses struct {
	Addr     []string
	Password string
	DB       int64
}

func errHndlr(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func makeClient(address string, password string, db int64) *redis.Client {
	client, err := redis.DialTimeout("tcp", address, time.Duration(10)*time.Second)
	errHndlr(err)
	if password != "" {
		r := client.Cmd("AUTH", password)
		errHndlr(r.Err)
	}
	r := client.Cmd("SELECT", db)
	errHndlr(r.Err)
	return client
}

func getStats(client *redis.Client) (*RedisStats, error) {
	info, err := client.Cmd("INFO").Str()
	errHndlr(err)

  maxmem_slice, err := client.Cmd("CONFIG", "GET", "maxmemory").List()
	errHndlr(err)

	usedmem_regex, _ := regexp.Compile(`used_memory:(\d+)`)
	clientconn_regex, _ := regexp.Compile(`connected_clients:(\d+)`)
	iops_regex, _ := regexp.Compile(`instantaneous_ops_per_sec:(\d+)`)
	evictedkeys_regex, _ := regexp.Compile(`evicted_keys:(\d+)`)
	db_regex, _ := regexp.Compile(`(db\d+):keys=(\d+),expires=(\d+)`)

	usedmem, err := strconv.ParseInt(usedmem_regex.FindStringSubmatch(info)[1], 10, 64)
	maxmem, err := strconv.ParseInt(maxmem_slice[len(maxmem_slice)-1], 10, 64)
	freemem := maxmem - usedmem
	evictedkeys, err := strconv.ParseInt(evictedkeys_regex.FindStringSubmatch(info)[1], 10, 64)
	iops, err := strconv.ParseInt(iops_regex.FindStringSubmatch(info)[1], 10, 64)
	clientconn, err := strconv.ParseInt(clientconn_regex.FindStringSubmatch(info)[1], 10, 64)
	dbs := make(map[string]*DBKeyspace)
	for _, db := range db_regex.FindAllStringSubmatch(info, -1) {
		keys, _ := strconv.ParseInt(db[2], 10, 64)
		expires, _ := strconv.ParseInt(db[3], 10, 64)
		dbs[db[1]] = &DBKeyspace{
			Keys:    keys,
			Expires: expires,
		}
	}
	stats := &RedisStats{
		FreeMemory:             freemem,
		EvictedKeys:            evictedkeys,
		InstantaneousOpsPerSec: iops,
		ClientConnections:      clientconn,
		DBs:                    dbs,
	}
	return stats, err
}

func (a *Addresses) yield() []byte {
	serverInfo := make(map[string]*RedisStats)
	for _, addr := range a.Addr {
		client := makeClient(addr, a.Password, a.DB)
		defer client.Close()
		stats, err := getStats(client)
		if err != nil {
			panic(err)
		}
		serverInfo[addr] = stats
	}
	content, _ := json.Marshal(serverInfo)
	return content
}

func (a *Addresses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Write(a.yield())
}

func main() {
	password := flag.String("password", "", "Password for all redis instances")
	db := flag.Int("db", 0, "DB number")
	socket := flag.String("socket", "/tmp/redis-monitor.sock", "Socket to provide metrics over")
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	s := &http.Server{
		Handler: &Addresses{
			Addr:     flag.Args(),
			Password: *password,
			DB:       int64(*db),
		},
	}

	l, err := util.CreateSocket(*socket)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	defer profile.Start(profile.MemProfile).Stop()

	if err != nil {
		panic(err)
	}

	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
