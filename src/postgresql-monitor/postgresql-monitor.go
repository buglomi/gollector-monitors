package main

import (
	"custerr"
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/bmizerany/pq"
	"net/http"
	"strings"
	"util"
)

type Attrs struct {
	Params       map[string]*string
	IgnoreErrors bool
}

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
var db_stats = map[string]string{
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

func (a *Attrs) yield() []byte {
	auth_string := ""
	results := map[string]int{}

	for key, value := range a.Params {
		if *value != "" {
			auth_string += key + "=" + *value + " "
		}
	}

	db, err := sql.Open("postgres", strings.Trim(auth_string, " \t"))

	if err != nil {
		return []byte("null")
	}

	defer db.Close()

	pg := &PGStat{DB: db}

	for key, value := range db_stats {
		results[key] = pg.getCount(value, a.IgnoreErrors)
	}

	content, err := json.Marshal(results)

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
	params := map[string]*string{}

	for key, value := range db_params {
		params[key] = flag.String(key, value[0], value[1])
	}

	ignore_errors := flag.Bool("ignore-errors", false, "Ignore processing errors while pulling stats -- nice if you don't have some relations")
	socket := flag.String("socket", "/tmp/postgresql-monitor.sock", "Socket to serve metrics on")

	flag.Parse()

	s := &http.Server{
		Handler: &Attrs{
			Params:       params,
			IgnoreErrors: *ignore_errors,
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
