package dblibsql

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	golibsql "github.com/tursodatabase/go-libsql"
)

type Turso struct {
	name          string
	url           string
	authToken     string
	encryptionKey string
	syncInterval  time.Duration

	db *sql.DB
}

func NewDatabase(name, url, token string, syncInterval time.Duration) Turso {
	return Turso{
		name:         name,
		url:          url,
		authToken:    token,
		syncInterval: syncInterval,
	}
}

func (d *Turso) ConnectEmbedded() (*sql.DB, error) {
	dbPath := filepath.Join("./", d.name)

	connector, err := golibsql.NewEmbeddedReplicaConnector(
		dbPath,
		d.url,
		golibsql.WithAuthToken(d.authToken),
		golibsql.WithSyncInterval(d.syncInterval),
	)

	if err != nil {
		fmt.Println("Error creating connector:", err)
		return nil, err
	}

	db := sql.OpenDB(connector)

	db.Ping() // check connectivity

	d.db = db
	return d.db, nil
}
