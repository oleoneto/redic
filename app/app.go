package app

import (
	"github.com/oleoneto/redic/app/controllers"
	"github.com/oleoneto/redic/app/domain/external"
	"github.com/oleoneto/redic/app/repositories"
	dbsql "github.com/oleoneto/redic/app/repositories/sql"
)

// Dependencies as per their interfaces
var (
	DatabaseEngine external.SqlEngineProtocol

	DictionaryRepository repositories.DictionaryRepository

	// Controllers
	DictionaryController controllers.DictionaryController

	NilValidator = func(any) map[string][]string { return map[string][]string{} }
)

func New(databaseOptions external.DBConnectOptions) {
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
