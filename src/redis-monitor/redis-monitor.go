package main

import (
	"encoding/json"
	"flag"
	"github.com/go-redis/redis"
	"github.com/gollector/gollector-monitors/src/util"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

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

func makeClient(address string, password string, db int64) *redis.Client {
	o := &redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	}
	client := redis.NewTCPClient(o)
	return client
}

func getStats(client *redis.Client) (*RedisStats, error) {
	info, err := client.Info().Result()
	if err != nil {
		return &RedisStats{}, err
	}

	usedmem_regex, _ := regexp.Compile(`used_memory:(\d+)`)
	clientconn_regex, _ := regexp.Compile(`connected_clients:(\d+)`)
	iops_regex, _ := regexp.Compile(`instantaneous_ops_per_sec:(\d+)`)
	evictedkeys_regex, _ := regexp.Compile(`evicted_keys:(\d+)`)
	db_regex, _ := regexp.Compile(`(db\d+):keys=(\d+),expires=(\d+)`)

	usedmem, err := strconv.ParseInt(usedmem_regex.FindStringSubmatch(info)[1], 10, 64)
	maxmem, err := strconv.ParseInt(client.ConfigGet("maxmemory").Val()[1].(string), 10, 64)
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
	err = client.Close()
	return stats, err
}

func (a *Addresses) yield() []byte {
	serverInfo := make(map[string]*RedisStats)
	for _, addr := range a.Addr {
		client := makeClient(addr, a.Password, a.DB)
		var stats *RedisStats
		stats, _ = getStats(client)
		serverInfo[addr] = stats
		defer client.Close()
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

	if err != nil {
		panic(err)
	}

	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
