package core

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Interfaces
type (
	LoaderFunc func(string) ([]fs.DirEntry, error)
	ReaderFunc func(string) ([]byte, error)
	ParserFunc func([]byte) map[string]DictEntry

	WordNet struct {
		loader LoaderFunc
		reader ReaderFunc
		parser ParserFunc
	}
)

func NewWordNet(l LoaderFunc, r ReaderFunc, p ParserFunc) *WordNet {
	return &WordNet{
		loader: l,
		reader: r,
		parser: p,
	}
}

func (wn *WordNet) LoadFiles(ctx context.Context, dir string) []fs.DirEntry {
	files, err := wn.loader(dir)
	if err != nil {
		return nil
	}

	return files
}

func (wn *WordNet) ParseContent(ctx context.Context, dir string, files []fs.DirEntry) error {
	for _, de := range files {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".yaml") {
			logrus.Errorln(de.Name(), "Error: Skipping invalid YAML file.")
			continue
		}

		path, _ := filepath.Abs(filepath.Join(dir, de.Name()))

		contents, err := wn.reader(path)
		if err != nil {
			logrus.Errorln(de.Name(), "Error:", err)
			continue
		}

		go wn.parser(contents)
	}

	return nil
}
