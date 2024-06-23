package external

type SQLAdapter string

const (
	PostgreSQLAdapter SQLAdapter = "postgresql"
	SQLite3Adapter    SQLAdapter = "sqlite3"
)

type DBConnectOptions struct {
	DB SqlEngineProtocol

	Adapter        SQLAdapter // [postgresql, sqlite3]
	DSN            string     // i.e postgresql://user:pass@host:5432/redic
	Filename       string     // i.e "redic.sqlite"
	VerboseLogging bool
}
