package protocols

import (
	"context"

	"github.com/oleoneto/redic/app/domain/types"
)

type DictionaryBackend interface {
	IndexWords(context.Context) error
	NewWords(context.Context, []types.NewWordInput) error
	AddWordDefinitions(context.Context, types.UpdateDefinitionInput) (types.Definitions, error)
	GetWordDefinitions(context.Context, types.GetWordDefinitionsInput) (types.WordDefinitions, error)
	GetDescribedWords(context.Context, types.GetDescribedWordsInput) (types.DescribedWords, error)
	// GetRelatedWords(context.Context, types.GetRelatedWordsInput) (types.RelatedWords, error)
}
