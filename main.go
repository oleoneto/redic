package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/oleoneto/redic/db"
	"github.com/oleoneto/redic/pkg/core"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	// Channels
	signaler = make(chan map[string]core.DictEntry)

	// Database
	database db.SqlEngineProtocol

	// Loader + Reader
	loader core.LoaderFunc = os.ReadDir
	reader core.ReaderFunc = os.ReadFile

	// Parser
	wordParser core.ParserFunc = func(b []byte) map[string]core.DictEntry {
		var entries map[string]core.DictEntry

		err := yaml.Unmarshal(b, &entries)
		if err != nil {
			return nil
		}

		signaler <- entries

		return entries
	}

	// Options
	basePath     = "./wordnet/english"
	databaseName = "redic.db"

	// Logger
	_ = logrus.New()
)

func main() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	ParseAndSave(ctx)

	fmt.Println("Total elapsed time:", time.Since(start))
}

func ParseAndSave(ctx context.Context) {
	db.UseSQLite(databaseName)

	wordnet := core.NewWordNet(loader, reader, wordParser)

	files := wordnet.LoadFiles(ctx, basePath)

	err := wordnet.ParseContent(ctx, basePath, files)
	if err != nil {
		panic(err)
	}

	for i := range files {
		select {
		case <-ctx.Done():
			fmt.Println(ctx.Err())
			os.Exit(0)

		case entries := <-signaler:
			total := len(entries)
			for _, e := range entries {
				for _, word := range e.Words() {
					word.Save(ctx, database)
				}
			}

			fmt.Printf("#%d -  %d entries in %s\n", i+1, total, files[i].Name())
		}
	}
}
