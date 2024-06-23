package controllers

import (
	"context"
	"fmt"

	"github.com/oleoneto/redic/app/domain/external"
	"github.com/oleoneto/redic/pkg/helpers"
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
	repository external.WordRepositoryProtocol
	validate   func(any) map[string][]string
}

func NewDictionaryController(repository external.WordRepositoryProtocol, validatorFunc func(any) map[string][]string) DictionaryController {
	return DictionaryController{
		repository: repository,
		validate:   validatorFunc,
	}
}

func (ctr *DictionaryController) CreateWords(ctx context.Context, data []external.NewWordInput) error {
	if errs := ctr.validate(data); len(errs) != 0 {
		return fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	err := ctr.repository.NewWords(ctx, data)

	return err
}

// Append to or create definitions for a dictionary entry
func (ctr *DictionaryController) UpdateDefinition(ctx context.Context, data external.UpdateDefinitionInput) (external.AddDefinitionsOutput, error) {
	if errs := ctr.validate(data); len(errs) != 0 {
		return external.AddDefinitionsOutput{}, fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	res, err := ctr.repository.AddWordDefinitions(ctx, data)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Given a dictionary entry, search for its corresponding definitions.
//
// Example:
//
//	`here` (noun):
//	the present location;
//	this place; location, proximal pronoun; demonstrative pronoun, location; quantifier: demonstrative determiner, singular, proximal
func (ctr *DictionaryController) GetDefinition(ctx context.Context, data external.GetWordDefinitionsInput) (external.GetWordDefinitionsOutput, error) {
	if errs := ctr.validate(data); len(errs) != 0 {
		return external.GetWordDefinitionsOutput{}, fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	res, err := ctr.repository.GetWordDefinitions(ctx, data)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Given a definition or word context, search for any matching words.
func (ctr *DictionaryController) FindMatchingWords(ctx context.Context, data external.GetDescribedWordsInput) (external.GetDescribedWordsOutput, error) {
	if errs := ctr.validate(data); len(errs) != 0 {
		return external.GetDescribedWordsOutput{}, fmt.Errorf(`invalid data for %v`, helpers.GetCurrentFuncName())
	}

	res, err := ctr.repository.GetDescribedWords(ctx, data)
	if err != nil {
		return res, err
	}

	return res, nil
}
