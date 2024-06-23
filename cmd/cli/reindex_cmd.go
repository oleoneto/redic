package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"log"

	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/app/domain/external"
	pkgcore "github.com/oleoneto/redic/pkg/core"
	"github.com/spf13/cobra"
)

var CreateTablesCmd = &cobra.Command{
	Use:   "create-tables",
	Short: "Create database tables",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		applyMigrations := func(ctx context.Context) error {
			tx, terr := state.Database.BeginTx(ctx, nil)
			if terr != nil {
				return terr
			}

			// -------------------------------
			// Instructions
			// -------------------------------

			_words := `
			CREATE TABLE words (
				id INTEGER PRIMARY KEY,
				word VARCHAR NOT NULL,
				part_of_speech CHAR NOT NULL,
				ili VARCHAR(50) NOT NULL,
		
				CONSTRAINT unique_word UNIQUE (word, part_of_speech, ili)
			)
			`
			_word_index := `CREATE INDEX word_idx ON words(word, part_of_speech)`

			_definitions := `
			CREATE TABLE definitions (
					id INTEGER PRIMARY KEY,
					word_id VARCHAR(50) NOT NULL REFERENCES words(id) UNIQUE,
					definitions TEXT NOT NULL
			)`

			_dictionary := `CREATE VIRTUAL TABLE redic_ USING fts5(definitions, word_id)`

			var tables = []string{
				`DROP TABLE IF EXISTS redic_;`,
				`DROP TABLE IF EXISTS definitions;`,
				`DROP TABLE IF EXISTS words;`,
				_words,
				_word_index,
				_definitions,
				_dictionary,
			}

			for _, stmt := range tables {
				if _, err := tx.ExecContext(ctx, stmt); err != nil {
					tx.Rollback()
					return err
				}
			}

			return tx.Commit()
		}

		if err := applyMigrations(ctx); err != nil {
			log.Fatalln(err)
			return
		}
	},
}

var ReindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Reindex the wordnet database",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Minute)
		defer cancel()

		parser := pkgcore.DefaultParser(
			os.ReadDir,
			os.ReadFile,
		)

		files := parser.LoadFiles(
			ctx,
			state.Flags.Directory,
		)

		if len(files) == 0 {
			return
		}

		// NOTE: Ensure tables exist
		// state.Database.QueryContext(ctx, )

		var wordCounter int
		for entries := range parser.ParseFilesChan(ctx, state.Flags.Directory, files) {
			total := len(entries)

			for _, e := range entries {
				// TODO: Review arguments to function call
				if err := app.DictionaryController.CreateWords(ctx, []external.NewWordInput{}); err != nil {
					panic(err)
				}
				wordCounter += len(e.Words())
			}

			fmt.Printf("%d entries processed\n", total)
		}

		fmt.Printf("%d words indexed\n", wordCounter)

		// .................................
		// .................................
		// .................................

		/*
			for i, f := range files {
				select {
				case entries := <-state.Signaler:

					// Last item
					if i == len(files)-1 {
						close(state.Signaler)

						_, err := state.Database.Exec("INSERT INTO redic_ (definitions, word_id) SELECT definitions, word_id FROM definitions")
						if err != nil {
							fmt.Println("Failed to rebuild dictionary 😢")
							panic(err)
						}

						fmt.Println("Done rebuilding 🎉")
					}
				case <-ctx.Done():
					os.Exit(0)
				}
			}
		*/
	},
}

func init() {
	ReindexCmd.Flags().StringVarP(&state.Flags.Directory, "dictionary-directory", "d", state.Flags.Directory, "dictionary directory")
	ReindexCmd.Flags().StringVarP(&state.Flags.DatabaseName, "database-name", "n", state.Flags.DatabaseName, "database name")
	CreateTablesCmd.Flags().StringVarP(&state.Flags.DatabaseName, "database-name", "n", state.Flags.DatabaseName, "database name")

	ReindexCmd.MarkFlagRequired("dictionary-directory")

	switch state.Flags.Engine.String() {
	case "postgresql":
	default:
		ReindexCmd.MarkFlagRequired("database-name")
		CreateTablesCmd.MarkFlagRequired("database-name")
	}
}
