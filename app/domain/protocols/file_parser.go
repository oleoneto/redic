package protocols

import (
	"context"
	"io/fs"

	"github.com/oleoneto/redic/app/domain/types"
)

type (
	LoaderFunc  func(string) ([]fs.DirEntry, error)
	ReaderFunc  func(string) ([]byte, error)
	ParserFunc  func([]byte) types.DictEntry
	CaptureFunc func(*types.ParsedFile) error
)

type FileParserProtocol interface {
	LoadFiles(context.Context, string) []fs.DirEntry
	ParseFile(context.Context, string, fs.DirEntry) (*types.ParsedFile, error)
	ParseFiles(context.Context, string, []fs.DirEntry, CaptureFunc) ([]types.ParsedFile, error)
}
