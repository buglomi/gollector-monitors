package main

import (
	"conversions"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/vmihailenco/redis"
	"os"
	"strings"
)

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

func yield(host string, port string, password string, dbnum int) {
	client := redis.NewTCPClient(host+":"+port, password, int64(dbnum))
	defer client.Close()

	info_string := client.Info()

	if info_string.Err() != nil {
		os.Stderr.WriteString(info_string.Err().Error() + "\n")
		fmt.Println("{}")
		os.Exit(1)
	}

	info := parseInfo(info_string.Val())

	conversions.ConvertTypes(&info)
	content, err := json.Marshal(info)

	if err != nil {
		panic(err)
		fmt.Println("{}")
		os.Exit(1)
	}

	fmt.Println(string(content))
}

func main() {
	host := flag.String("host", "localhost", "Hostname of redis instance")
	port := flag.String("port", "6379", "Port of redis instance")
	password := flag.String("password", "", "Password to connect to redis instance")
	dbnum := flag.Int("dbnum", -1, "Database number")
	flag.Parse()

	if *host == "" || *port == "" {
		os.Stderr.WriteString("Please enter a valid host and port\n")
		fmt.Println("{}")
		os.Exit(1)
	}

	yield(*host, *port, *password, *dbnum)
}
