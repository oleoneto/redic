package parsers

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/oleoneto/redic/app/domain/protocols"
	"github.com/oleoneto/redic/app/domain/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	loader protocols.LoaderFunc
	reader protocols.ReaderFunc
}

var _ protocols.FileParserProtocol = (*Parser)(nil)

func DefaultParser(l protocols.LoaderFunc, r protocols.ReaderFunc) *Parser {
	return &Parser{loader: l, reader: r}
}

func (p *Parser) LoadFiles(ctx context.Context, dir string) []fs.DirEntry {
	files, err := p.loader(dir)
	if err != nil {
		return nil
	}
	return files
}

func (p *Parser) ParseFile(ctx context.Context, dir string, file fs.DirEntry) (*types.ParsedFile, error) {
	if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
		logrus.Errorln(file.Name(), "Error: Skipping invalid YAML file.")
		return nil, nil
	}

	path := filepath.Join(dir, file.Name())

	contents, err := p.reader(path)
	if err != nil {
		logrus.Errorln(file.Name(), "Error:", err)
		return nil, err
	}

	var pf map[string]types.DictEntry
	err = yaml.Unmarshal(contents, &pf)
	if err != nil {
		return nil, err
	}

	return &types.ParsedFile{Data: pf, Name: file.Name()}, nil
}

func (p *Parser) ParseFiles(ctx context.Context, dir string, files []fs.DirEntry, capturer protocols.CaptureFunc) ([]types.ParsedFile, error) {
	var entries = []types.ParsedFile{}

	for _, de := range files {
		parsedFile, err := p.ParseFile(ctx, dir, de)
		if err != nil || parsedFile == nil {
			return nil, err
		}

		entries = append(entries, *parsedFile)

		go capturer(parsedFile)
	}

	return entries, nil
}

func (p *Parser) ParseFilesChan(ctx context.Context, dir string, files []fs.DirEntry) <-chan map[string]types.DictEntry {
	dst := make(chan map[string]types.DictEntry)

	var next int
	var total = len(files)

	parsedFile, err := p.ParseFile(ctx, dir, files[next])
	if err != nil || parsedFile == nil {
		return nil
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return // prevent leak
			case dst <- parsedFile.Data:
				if next > total {
					close(dst)
					return // prevent leak
				}
				next++
			}
		}
	}()

	return dst
}
