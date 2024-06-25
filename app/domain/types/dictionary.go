package types

type PartOfSpeech string

type (
	NewWordInput struct {
		Word         string // i.e emerging
		PartOfSpeech string // i.e a
		Definition   string // i.e comming into existence
	}

	NewWordsOutput struct{}

	UpdateDefinitionInput struct {
		Word         string
		PartOfSpeech PartOfSpeech
		Definitions  string
	}

	Definitions struct{ Id string }

	GetWordDefinitionsInput struct {
		Word         string
		PartOfSpeech PartOfSpeech
		Verbatim     bool
	}

	WordDefinitions struct {
		Word        string `json:"word"`
		Definitions []struct {
			PartOfSpeech PartOfSpeech `json:"part_of_speech"`
			Definition   string       `json:"definition"`
		} `json:"definitions"`
	}

	GetDescribedWordsInput struct {
		Descriptions []string
		Verbatim     bool
	}

	DescribedWords struct {
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
