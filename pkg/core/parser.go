package core

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Interfaces
type (
	ParsedFile struct {
		Name string
		Data map[string]DictEntry
	}

	LoaderFunc  func(string) ([]fs.DirEntry, error)
	ReaderFunc  func(string) ([]byte, error)
	ParserFunc  func([]byte) DictEntry
	CaptureFunc func(*ParsedFile) error

	ParserProtocol interface {
		LoadFiles(context.Context, string) []fs.DirEntry
		ParseFile(context.Context, string, fs.DirEntry) (*ParsedFile, error)
		ParseFiles(context.Context, string, []fs.DirEntry, CaptureFunc) ([]ParsedFile, error)
	}

	Parser struct {
		loader LoaderFunc
		reader ReaderFunc
	}
)

var _ ParserProtocol = (*Parser)(nil)

func DefaultParser(l LoaderFunc, r ReaderFunc) *Parser {
	return &Parser{
		loader: l,
		reader: r,
	}
}

func (wn *Parser) LoadFiles(ctx context.Context, dir string) []fs.DirEntry {
	files, err := wn.loader(dir)
	if err != nil {
		return nil
	}
	return files
}

func (wn *Parser) ParseFile(ctx context.Context, dir string, file fs.DirEntry) (*ParsedFile, error) {
	if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
		logrus.Errorln(file.Name(), "Error: Skipping invalid YAML file.")
		return nil, nil
	}

	path, _ := filepath.Abs(filepath.Join(dir, file.Name()))

	contents, err := wn.reader(path)
	if err != nil {
		logrus.Errorln(file.Name(), "Error:", err)
		return nil, err
	}

	var pf map[string]DictEntry
	err = yaml.Unmarshal(contents, &pf)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{Data: pf, Name: file.Name()}, nil
}

func (wn *Parser) ParseFiles(ctx context.Context, dir string, files []fs.DirEntry, capturer CaptureFunc) ([]ParsedFile, error) {
	var entries = []ParsedFile{}

	for _, de := range files {
		parsedFile, err := wn.ParseFile(ctx, dir, de)
		if err != nil || parsedFile == nil {
			return nil, err
		}

		entries = append(entries, *parsedFile)

		go capturer(parsedFile)
	}

	return entries, nil
}
