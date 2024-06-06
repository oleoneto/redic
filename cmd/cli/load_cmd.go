package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oleoneto/redic/db"
	pkgcore "github.com/oleoneto/redic/pkg/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the wordnet database",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
		defer cancel()

		state.Database, _ = db.UseSQLite(state.Flags.DatabaseName)

		// Migrate tables
		migrations := "./db/migrations"
		if err := applyMigrations(ctx, migrations, state.Database); err != nil {
			panic(err)
		}

		wordnet := pkgcore.NewWordNet(os.ReadDir, os.ReadFile)

		files := wordnet.LoadFiles(ctx, state.Flags.Directory)

		if err := wordnet.ParseContent(ctx, state.Flags.Directory, files, wordParser); err != nil {
			panic(err)
		}

		for i := range files {
			select {
			case entries := <-state.Signaler:
				total := len(entries)
				for _, e := range entries {
					for _, word := range e.Words() {
						word.Save(ctx, state.Database)
					}
				}
				fmt.Printf("#%d -  %d entries in %s\n", i+1, total, files[i].Name())

			case <-ctx.Done():
				fmt.Println(ctx.Err())
				os.Exit(0)
			}
		}
	},
}

func init() {
	BuildCmd.Flags().StringVarP(&state.Flags.Directory, "directory", "d", state.Flags.Directory, "")
	BuildCmd.Flags().StringVarP(&state.Flags.DatabaseName, "database-name", "n", state.Flags.DatabaseName, "")

	BuildCmd.MarkFlagRequired("directory")
	BuildCmd.MarkFlagRequired("database-name")
}

func applyMigrations(ctx context.Context, dir string, db db.SqlEngineProtocol) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	type changes struct {
		Up   []string `yaml:"up"`
		Down []string `yaml:"down"`
	}

	type migration struct {
		Changes changes `yaml:"changes,omitempty" json:"-"`
	}

	for _, de := range files {
		path, _ := filepath.Abs(filepath.Join(dir, de.Name()))

		contents, err := os.ReadFile(path)
		if err != nil {
			logrus.Errorln(de.Name(), "Error:", err)
			return err
		}

		var m migration
		yaml.Unmarshal(contents, &m)

		tx, terr := db.BeginTx(ctx, nil)
		if terr != nil {
			return err
		}

		for _, inst := range m.Changes.Up {
			if _, err := tx.ExecContext(ctx, inst); err != nil {
				tx.Rollback()
				return err
			}
		}

		tx.Commit()
	}

	return nil
}

func wordParser(b []byte) map[string]pkgcore.DictEntry {
	var entries map[string]pkgcore.DictEntry

	err := yaml.Unmarshal(b, &entries)
	if err != nil {
		return nil
	}

	state.Signaler <- entries

	return entries
}
