package protocols

import (
	"context"

	"github.com/oleoneto/redic/app/domain/types"
)

type DictionaryBackend interface {
	IndexWords(context.Context) error
	NewWords(context.Context, []types.NewWordInput) error
	// AddWordDefinitions(context.Context, types.UpdateDefinitionInput) (types.Definitions, error)
	GetWordExplanation(context.Context, types.GetWordDefinitionsInput) (types.WordDefinitions, error)
	SearchWords(context.Context, types.GetDescribedWordsInput) (types.WordMatches, error)
	// GetRelatedWords(context.Context, types.GetRelatedWordsInput) (types.RelatedWords, error)
}
