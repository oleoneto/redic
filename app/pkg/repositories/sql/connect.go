package dbsql

import (
	"fmt"
	"os"

	"github.com/oleoneto/redic/app/domain/protocols"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

func ConnectDatabase(options protocols.DBConnectOptions) (protocols.SqlBackend, error) {
	switch options.Adapter {
	case protocols.PostgreSQLAdapter:
		if options.DSN == "" {
			return nil, fmt.Errorf("database-url not set")
		}

		db, err := UsePG(options.DSN)
		if err != nil {
			return nil, err
		}

		if db == nil {
			return db, fmt.Errorf("database connection failed")
		}

		return db, nil
	case protocols.SQLite3Adapter:
		db, err := UseSQLite(options.Filename)
		if err != nil {
			return nil, err
		}

		if db == nil {
			return db, fmt.Errorf("database connection failed")
		}

		if options.VerboseLogging {
			loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
			db = sqldblogger.OpenDriver(options.Filename, db.Driver(), loggerAdapter)

			if db == nil {
				return db, fmt.Errorf("database logger failed")
			}
		}

		db.Exec("PRAGMA journal_mode=WAL")

		return db, nil
	}
	return nil, fmt.Errorf("database adapter not set")
}
