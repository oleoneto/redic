package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"log"

	"github.com/mitchellh/go-homedir"
	"github.com/oleoneto/go-toolkit/files"
	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/app/domain/types"
	"github.com/oleoneto/redic/app/pkg/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var resetTables bool
var repopulateDatabase bool
var copyDefaultDatabase bool

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CLI by creating its required configuration files",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		state.Flags.HomeDirectory, err = homedir.Dir()
		if err != nil {
			log.Fatalln(err)
		}

		state.Flags.ConfigDir.Create(files.FileGenerator{}, state.Flags.HomeDirectory)

		f := filepath.Join(state.Flags.HomeDirectory, state.Flags.ConfigDir.Name, "data", dbfile)
		viper.Set("database.path", f)
		viper.WriteConfig()

		if copyDefaultDatabase {
			CopyDatabase(cmd, args)
			return
		}

		if resetTables || repopulateDatabase {
			state.ConnectDatabase(cmd, args)
		}

		CreateTables(cmd, args)

		PopulateTables(cmd, args)
	},
}

func CopyDatabase(cmd *cobra.Command, args []string) {
	// if !copyDefaultDatabase {
	// 	return
	// }

	// Copy embeded dictionary to local filesystem

	f := filepath.Join("data", dbfile)
	dictionary, err := virtualFS.Open(f)
	if err != nil {
		log.Fatalln(err)
	}

	dstDir := filepath.Join(state.Flags.HomeDirectory, state.Flags.ConfigDir.Name, "data", dbfile)
	dstFile, err := os.Create(dstDir)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := io.Copy(dstFile, dictionary); err != nil {
		log.Fatalln(err)
	}
}

func CreateTables(cmd *cobra.Command, args []string) {
	if !resetTables {
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	createTables := func(ctx context.Context) error {
		tx, terr := state.Database.BeginTx(ctx, nil)
		if terr != nil {
			return terr
		}

		f := filepath.Join("data", "redic.sql")
		b, err := virtualFS.ReadFile(f)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(b)); err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}

	if err := createTables(ctx); err != nil {
		log.Fatalln(err)
		return
	}
}

func PopulateTables(cmd *cobra.Command, args []string) {
	if !repopulateDatabase {
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Minute)
	defer cancel()

	parser := parsers.DefaultParser(
		virtualFS.ReadDir,
		virtualFS.ReadFile,
	)

	dictDirectory := filepath.Join("data", "english")
	files := parser.LoadFiles(
		ctx,
		dictDirectory,
	)

	if len(files) == 0 {
		return
	}

	defer app.DictionaryController.IndexWords(ctx)

	fmt.Printf("%d files to process\n", len(files))

	ch := make(chan types.DictFile)
	parser.ParseFiles(ctx, dictDirectory, files, func(pf *types.ParsedFile) error {
		ch <- pf.Data
		return nil
	})

	for _, i := range files {
		select {
		case file := <-ch:
			fmt.Println("Processsing", i.Name())

			for _, f := range file {
				if err := app.DictionaryController.CreateWords(ctx, f.Words()); err != nil {
					log.Fatalln(err)
				}
			}
		case <-ctx.Done():
			fmt.Printf("Done processing all %d files\n", len(files))
		}
	}
}

func init() {
	InitCmd.Flags().BoolVar(&resetTables, "reset-tables", resetTables, "")
	InitCmd.Flags().BoolVar(&repopulateDatabase, "repopulate", repopulateDatabase, "")
	InitCmd.Flags().BoolVar(&copyDefaultDatabase, "copy-db", copyDefaultDatabase, "")
}
