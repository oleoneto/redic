package db

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/mattn/go-sqlite3"
)

type SqlEngineProtocol interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func init() {
	var regex = func(re, s string) (bool, error) { return regexp.MatchString(re, s) }

	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		},
	)
}
