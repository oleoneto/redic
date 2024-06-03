package parsers

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

type WordNet struct{}

func (wn *WordNet) LoadFiles(ctx context.Context, dir string, loader func(dir string) ([]fs.DirEntry, error)) []fs.DirEntry {
	files, err := loader(dir)
	if err != nil {
		return nil
	}

	return files
}

func (wn *WordNet) ParseContent(ctx context.Context, dir string, files []fs.DirEntry, reader func(filename string) ([]byte, error), processor func([]byte)) error {
	for _, de := range files {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".yaml") {
			return nil
		}

		path, _ := filepath.Abs(filepath.Join(dir, de.Name()))

		contents, err := reader(path)
		if err != nil {
			return err
		}

		go processor(contents)
	}

	return nil
}
