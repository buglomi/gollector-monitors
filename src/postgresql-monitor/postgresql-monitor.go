package main

import (
	"custerr"
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/bmizerany/pq"
	"os"
	"strings"
)

type PGStat struct {
	DB *sql.DB
}

var db_params = map[string][2]string{
	"user":     [2]string{"postgres", "The username to access postgresql"},
	"password": [2]string{"", "The password (if any) to access postgresql"},
	"dbname":   [2]string{"template1", "The database name."},
	"sslmode":  [2]string{"disable", "SSL Mode. Options are disable, required (no verification), and verify-full"},
	"host":     [2]string{"localhost", "The host. Start with a / to use a unix socket"},
	"port":     [2]string{"", "The port, only useful for TCP connections."},
}

// this gets transformed into ints based on the count of the value.
// kinda ugly but it's better than a bunch of funcs that do the same thing.
var db_stats = map[string]interface{}{
	"matviews":              "pg_matviews",
	"locks":                 "pg_locks",
	"cursors":               "pg_cursors",
	"prepared_statements":   "pg_prepared_statements",
	"prepared_transactions": "pg_prepared_xacts",
}

func (pg *PGStat) getCount(table_name string, ignore_errors bool) int {
	var val int

	if err := pg.DB.QueryRow("select count(*) from \"" + table_name + "\"").Scan(&val); err != nil {
		if ignore_errors {
			return 0
		}

		custerr.Fatal("Error trying to query " + table_name + ": " + err.Error())
	}

	return val
}

func yield(params map[string]*string, ignore_errors bool) {
	auth_string := ""

	for key, value := range params {
		if *value != "" {
			auth_string += key + "=" + *value + " "
		}
	}

	db, err := sql.Open("postgres", strings.Trim(auth_string, " \t"))

	if err != nil {
		custerr.Fatal("Error connecting to postgresql database: " + err.Error())
	}

	defer db.Close()

	pg := &PGStat{DB: db}

	for key, value := range db_stats {
		db_stats[key] = pg.getCount(value.(string), ignore_errors)
	}

	content, err := json.Marshal(db_stats)

	if err != nil {
		custerr.Fatal("Error marshalling content: " + err.Error())
	}

	os.Stdout.Write(content)
	os.Exit(0)
}

func main() {
	params := map[string]*string{}

	for key, value := range db_params {
		params[key] = flag.String(key, value[0], value[1])
	}

	ignore_errors := flag.Bool("ignore-errors", false, "Ignore processing errors while pulling stats -- nice if you don't have some relations")

	flag.Parse()

	yield(params, *ignore_errors)
}
