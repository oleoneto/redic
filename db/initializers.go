package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func UsePG(dsn string) *sql.DB {
	d, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	return d
}

func UseSQLite(dbname string) *sql.DB {
	d, err := sql.Open("sqlite3_extended", fmt.Sprintf("%s?_fk=on&_ignore_check_constraints=off&_journal=WAL&_cslike=off", dbname))
	if err != nil {
		panic(err)
	}

	return d
}
