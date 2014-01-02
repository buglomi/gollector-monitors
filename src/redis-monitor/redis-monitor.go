package main

import (
	"encoding/json"
	//"flag"
	"fmt"
	"github.com/vmihailenco/redis"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FIXME move this to a generic lib later
var CONV_TABLE = map[string]func(float64) float64{
	"K": func(f float64) float64 {
		return f * 1024
	},
	"M": func(f float64) float64 {
		return f * 1024 * 1024
	},
	"G": func(f float64) float64 {
		return f * 1024 * 1024 * 1024
	},
	"T": func(f float64) float64 {
		return f * 1024 * 1024 * 1024 * 1024
	},
	"P": func(f float64) float64 {
		return f * 1024 * 1024 * 1024 * 1024 * 1024
	},
}

func convertTypes(info *map[string]interface{}) {
	for key, value := range *info {
		strval := value.(string)
		isnum, _ := regexp.MatchString(`\A-?\d+\z`, strval)

		if isnum {
			i, err := strconv.ParseInt(strval, 10, 64)

			if err == nil {
				(*info)[key] = i
				continue
			}
		}

		isfloat, _ := regexp.MatchString(`\A-?\d+\.\d+\z`, strval)

		if isfloat {
			f, err := strconv.ParseFloat(strval, 64)

			if err == nil {
				(*info)[key] = f
				continue
			}
		}

		is_si, _ := regexp.MatchString(`\A-?\d+(?:\.\d+)?[kKmMgGtTpP]\z`, strval)

		if is_si {
			val := strval[0 : len(strval)-1]
			unit := strval[len(strval)-1 : len(strval)]

			f, err := strconv.ParseFloat(val, 64)

			if err == nil {
				fmt.Println(key, f, unit, CONV_TABLE[strings.ToUpper(unit)](f))

				(*info)[key] = CONV_TABLE[strings.ToUpper(unit)](f)
				continue
			}
		}
	}
}

func main() {
	info := map[string]interface{}{}

	client := redis.NewTCPClient("localhost:6379", "", int64(-1))
	defer client.Close()

	info_string := client.Info()

	if info_string.Err() != nil {
		os.Stderr.Write([]byte(info_string.Err().Error()))
		os.Exit(1)
	}

	lines := strings.Split(info_string.Val(), "\r\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && len(strings.Trim(line, " \t")) != 0 {
			values := strings.SplitN(line, ":", 2)
			info[values[0]] = values[1]
		}
	}

	convertTypes(&info)
	content, err := json.Marshal(info)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
