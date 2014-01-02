package conversions

import (
	"regexp"
	"strconv"
	"strings"
)

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

func ConvertTypes(info *map[string]interface{}) {
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
				(*info)[key] = CONV_TABLE[strings.ToUpper(unit)](f)
				continue
			}
		}
	}
}
