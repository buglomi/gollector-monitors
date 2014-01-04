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

type JSONOut struct {
	MaterializedViews int `json:matviews`
	Locks             int `json:locks`
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

func (pg *PGStat) getLocks() int {
	return pg.getCount("pg_locks")
}

func (pg *PGStat) getMaterializedViews() int {
	return pg.getCount("pg_matviews")
}

func main() {
	db, err := sql.Open("postgres", "user=erikh dbname=template1 sslmode=disable")

	if err != nil {
		os.Stderr.WriteString("Error connecting to postgresql database: " + err.Error())
		os.Exit(1)
	}

	defer db.Close()

	pg := &PGStat{DB: db}

	json_out := JSONOut{
		MaterializedViews: pg.getMaterializedViews(),
		Locks:             pg.getLocks(),
	}

	content, err := json.Marshal(json_out)

	if err != nil {
		os.Stderr.WriteString("Error marshalling content: " + err.Error())
		os.Exit(1)
	}

	os.Stdout.Write(content)
	os.Exit(0)
}
