package types

import "encoding/json"

type PartOfSpeech string

type (
	NewWordInput struct {
		Word         string // i.e emerging
		PartOfSpeech string // i.e a
		Definition   string // i.e comming into existence
		Explicit     bool
	}

	NewWordsOutput struct{}

	UpdateDefinitionInput struct {
		Word         string
		PartOfSpeech PartOfSpeech
		Definitions  string
	}

	Definitions struct{ Id string }

	GetWordDefinitionsInput struct {
		Word         string       `json:"word"`
		PartOfSpeech PartOfSpeech `json:"part_of_speech"`
		Verbatim     bool         `json:"verbatim"`
	}

	Definition struct {
		PartOfSpeech PartOfSpeech `json:"part_of_speech"`
		Definition   string       `json:"text"`
		Explicit     bool         `json:"explicit,omitempty"`
	}

	WordDefinitions struct {
		Word        string       `json:"word"`
		Definitions []Definition `json:"definitions"`
	}

	GetDescribedWordsInput struct {
		Cursor          string       `json:"cursor_id"`
		Tokens          string       `json:"description"`
		PartOfSpeech    PartOfSpeech `json:"part_of_speech"`
		IncludeExplicit bool         `json:"include_explicit"`
	}

	MatchingWord struct {
		Id           int          `json:"id"`
		Word         string       `json:"word"`
		PartOfSpeech PartOfSpeech `json:"part_of_speech"`
		Definition   string       `json:"definition"`
		Explicit     bool         `json:"explicit,omitempty"`
	}

	WordMatches struct {
		Cursor               string         `json:"cursor_id,omitempty"`
		ProvidedDescriptions string         `json:"query,omitempty"`
		MatchingWords        []MatchingWord `json:"matching_words"`
	}
)

const (
	Adjective1 PartOfSpeech = "a"
	Adjective2 PartOfSpeech = "s"
	Adverb     PartOfSpeech = "r"
	Noun       PartOfSpeech = "n"
	Verb       PartOfSpeech = "v"
	ALL        PartOfSpeech = "*"
)

func (p *PartOfSpeech) MarshalJSON() ([]byte, error) {
	type P string
	return json.Marshal(P(p.Raw()))
}

func (p PartOfSpeech) Raw() string {
	switch p {
	case Adjective1, Adjective2:
		return "adjective"
	case Adverb:
		return "adverb"
	case Noun:
		return "noun"
	case Verb:
		return "verb"
	}

	return "*"
}
