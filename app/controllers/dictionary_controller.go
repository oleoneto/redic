package controllers

import (
	"context"
	"fmt"

	"github.com/oleoneto/redic/app/domain/protocols"
	"github.com/oleoneto/redic/app/domain/types"
	"github.com/oleoneto/redic/app/pkg/helpers"
)

type SearchMode int

const (
	// Given a dictionary entry, search for its corresponding definitions.
	Define SearchMode = iota

	// Given a definition or word context, search for any matching words.
	Lookup SearchMode = iota
)

type DictionarySearch struct {
	// The word, words, or definitions one wishes to search for.
	Input string

	// Used to determine how searches should be performed.
	Mode SearchMode

	// When true, searches will be performed using exactly the provided input.
	// No similarity searches or approximations will be used.
	Verbatim bool
}

type DictionaryController struct {
	repository protocols.DictionaryBackend
	validate   func(any) map[string][]string
}

func NewDictionaryController(repository protocols.DictionaryBackend, validatorFunc func(any) map[string][]string) DictionaryController {
	return DictionaryController{
		repository: repository,
		validate:   validatorFunc,
	}
}

// Append to or create words + definitions to the dictionary
func (ctr *DictionaryController) CreateWords(ctx context.Context, data []types.NewWordInput) error {
	if errs := ctr.validate(data); len(errs) != 0 {
		return fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	err := ctr.repository.NewWords(ctx, data)

	return err
}

// Given a dictionary entry, search for its corresponding definitions.
//
// Example:
//
//	`here` (noun):
//	the present location;
//	this place; location, proximal pronoun; demonstrative pronoun, location; quantifier: demonstrative determiner, singular, proximal
func (ctr *DictionaryController) GetDefinition(ctx context.Context, data types.GetWordDefinitionsInput) (types.WordDefinitions, error) {
	if errs := ctr.validate(data); len(errs) != 0 {
		return types.WordDefinitions{}, fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	res, err := ctr.repository.GetWordExplanation(ctx, data)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Given a definition or word context, search for any matching words.
func (ctr *DictionaryController) FindMatchingWords(ctx context.Context, data types.GetDescribedWordsInput) (types.WordMatches, error) {
	if errs := ctr.validate(data); len(errs) != 0 {
		return types.WordMatches{}, fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	fmt.Println(helpers.GetCurrentFuncName(), data)

	res, err := ctr.repository.SearchWords(ctx, data)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (ctr *DictionaryController) IndexWords(ctx context.Context) error {
	return ctr.repository.IndexWords(ctx)
}
