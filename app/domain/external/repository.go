package external

import "context"

type PartOfSpeech string

type (
	NewWordInput struct {
		Word         string   // i.e emerging
		PartOfSpeech string   // i.e a
		Definitions  []string // i.e comming into existence
		Examples     any      // i.e an emergent republic

		EntryCode string
	}

	NewWordsOutput struct{}

	UpdateDefinitionInput struct {
		Word         string
		PartOfSpeech PartOfSpeech
		Definitions  []string
	}

	AddDefinitionsOutput struct{ Id string }

	GetWordDefinitionsInput struct {
		Word         string
		PartOfSpeech PartOfSpeech
		Verbatim     bool
	}

	GetWordDefinitionsOutput struct {
		Word        string
		Definitions []struct {
			PartOfSpeech PartOfSpeech
			Definition   string
		}
	}

	GetDescribedWordsInput struct {
		Descriptions []string
		Verbatim     bool
	}

	GetDescribedWordsOutput struct {
		ProvidedDescriptions []string
		MatchingWords        []struct {
			Id           string
			Word         string
			PartOfSpeech PartOfSpeech
			Definition   string
		}
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

type WordRepositoryProtocol interface {
	NewWords(context.Context, []NewWordInput) error
	AddWordDefinitions(context.Context, UpdateDefinitionInput) (AddDefinitionsOutput, error)
	GetWordDefinitions(context.Context, GetWordDefinitionsInput) (GetWordDefinitionsOutput, error)
	GetDescribedWords(context.Context, GetDescribedWordsInput) (GetDescribedWordsOutput, error)
}
