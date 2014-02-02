package main

import (
	"conversions"
	"custerr"
	"encoding/json"
	"flag"
	"github.com/vmihailenco/redis"
	"net/http"
	"strings"
	"util"
)

type Attrs struct {
	Host     string
	Port     string
	Password string
	DBNum    int
}

func parseInfo(info_string string) map[string]interface{} {
	info := map[string]interface{}{}

	lines := strings.Split(info_string, "\r\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && len(strings.Trim(line, " \t")) != 0 {
			values := strings.SplitN(line, ":", 2)
			info[values[0]] = values[1]
		}
	}

	return info
}

func (a *Attrs) yield() []byte {
	client := redis.NewTCPClient(a.Host+":"+a.Port, a.Password, int64(a.DBNum))
	defer client.Close()

	info_string := client.Info()

	if info_string.Err() != nil {
		return []byte("null")
	}

	info := parseInfo(info_string.Val())

	conversions.ConvertTypes(&info)
	content, err := json.Marshal(info)

	if err != nil {
		return []byte("null")
	}

	return content
}

func (a *Attrs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Write(a.yield())
}

func main() {
	host := flag.String("host", "localhost", "Hostname of redis instance")
	port := flag.String("port", "6379", "Port of redis instance")
	password := flag.String("password", "", "Password to connect to redis instance")
	dbnum := flag.Int("dbnum", -1, "Database number")
	socket := flag.String("socket", "/tmp/redis-monitor.sock", "Socket to provide metrics over")

	flag.Parse()

	if *host == "" || *port == "" {
		custerr.Fatal("Please enter a valid host and port\n")
	}

	s := &http.Server{
		Handler: &Attrs{
			Host:     *host,
			Port:     *port,
			Password: *password,
			DBNum:    *dbnum,
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
