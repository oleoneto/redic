package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/oleoneto/redic/pkg/core"
	"github.com/oleoneto/redic/pkg/parsers"
	"gopkg.in/yaml.v3"
)

func main() {
	start := time.Now()

	wordnet := parsers.WordNet{}

	var out = make(chan map[string]core.WordEntry)

	path := "./wordnet/english"

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	files := wordnet.LoadFiles(ctx, path, os.ReadDir)

	wordnet.ParseContent(ctx, path, files, os.ReadFile, func(b []byte) {
		var entries map[string]core.WordEntry

		err := yaml.Unmarshal(b, &entries)
		if err != nil {
			return
		}

		out <- entries
	})

	var words []core.Word

	for i := 0; i < len(files); i++ {
		select {
		case entries := <-out:
			for _, e := range entries {
				words = append(words, e.Words()...)
			}

			fmt.Println(len(words), "words from", len(entries), "entries")
		case <-ctx.Done():
			fmt.Println("Done")
		}
	}

	fmt.Println("Elapsed time:", time.Since(start))
}
