package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/bmizerany/pq"
	"os"
)

type PGStat struct {
	DB *sql.DB
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

func (pg *PGStat) getCount(table_name string) int {
	var val int

	err := pg.DB.QueryRow("select count(*) from \"" + table_name + "\"").Scan(&val)

	if err != nil {
		os.Stderr.WriteString("Error trying to query " + table_name + ": " + err.Error())
		os.Exit(1)
	}

	return val
}

func main() {
	db, err := sql.Open("postgres", "user=erikh dbname=template1 sslmode=disable")

	if err != nil {
		os.Stderr.WriteString("Error connecting to postgresql database: " + err.Error())
		os.Exit(1)
	}

	defer db.Close()

	pg := &PGStat{DB: db}

	for key, value := range db_stats {
		db_stats[key] = pg.getCount(value.(string))
	}

	content, err := json.Marshal(db_stats)

	if err != nil {
		os.Stderr.WriteString("Error marshalling content: " + err.Error())
		os.Exit(1)
	}

	os.Stdout.Write(content)
	os.Exit(0)
}
