package dbsql

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mattn/go-sqlite3"
)

func UsePG(dsn string) (*sql.DB, error) {
	d, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func UseSQLite(dbname string) (*sql.DB, error) {
	var regex = func(re, s string) (bool, error) { return regexp.MatchString(re, s) }

	sql.Register("sqlite3_ext",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		},
	)

	d, err := sql.Open("sqlite3_ext", fmt.Sprintf("%s?fts5=on&_fk=on&_ignore_check_constraints=off&_journal=WAL&_cslike=off", dbname))
	if err != nil {
		return nil, err
	}

	return d, nil
}
