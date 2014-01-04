package main

import (
	"database/sql"
	"fmt"
	_ "github.com/bmizerany/pq"
	"os"
)

type PGStat struct {
	DB *sql.DB
}

func (pg *PGStat) getLocks() int {
	var val int

	err := pg.DB.QueryRow("select count(*) from pg_locks").Scan(&val)

	if err != nil {
		os.Stderr.WriteString("Error trying to query pg_locks: " + err.Error())
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
	fmt.Println(pg.getLocks())
}
