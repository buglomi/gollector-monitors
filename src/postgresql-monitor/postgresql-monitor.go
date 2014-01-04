package main

import (
	"custerr"
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

	if err := pg.DB.QueryRow("select count(*) from \"" + table_name + "\"").Scan(&val); err != nil {
		custerr.Fatal("Error trying to query " + table_name + ": " + err.Error())
	}

	return val
}

func main() {
	db, err := sql.Open("postgres", "user=erikh dbname=template1 sslmode=disable")

	if err != nil {
		custerr.Fatal("Error connecting to postgresql database: " + err.Error())
	}

	defer db.Close()

	pg := &PGStat{DB: db}

	for key, value := range db_stats {
		db_stats[key] = pg.getCount(value.(string))
	}

	content, err := json.Marshal(db_stats)

	if err != nil {
		custerr.Fatal("Error marshalling content: " + err.Error())
	}

	os.Stdout.Write(content)
	os.Exit(0)
}
