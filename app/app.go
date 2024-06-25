package app

import (
	"github.com/oleoneto/redic/app/controllers"
	"github.com/oleoneto/redic/app/domain/protocols"
	"github.com/oleoneto/redic/app/pkg/repositories"
	dbsql "github.com/oleoneto/redic/app/pkg/repositories/sql"
)

var CommitHash = "unset"

// Dependencies as per their interfaces
var (
	DatabaseEngine protocols.SqlBackend

	DictionaryRepository repositories.DictionaryRepository

	// Controllers
	DictionaryController controllers.DictionaryController

	NilValidator = func(any) map[string][]string { return map[string][]string{} }
)

func New(databaseOptions protocols.DBConnectOptions) {
	if databaseOptions.DB != nil {
		DatabaseEngine = databaseOptions.DB
	} else {
		db, err := dbsql.ConnectDatabase(databaseOptions)
		if err != nil {
			panic(err)
		}

		DatabaseEngine = db
	}

	DictionaryRepository = *repositories.NewDictionaryRepository(DatabaseEngine)

	DictionaryController = controllers.NewDictionaryController(
		&DictionaryRepository,
		NilValidator,
	)
}
