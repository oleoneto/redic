package core

/**
00003552-s:
  definition:
  - coming into existence
  example:
  - an emergent republic
  ili: i10
  members:
  - emergent
  - emerging
  partOfSpeech: s
  similar:
  - 00003356-a
---
02057872-a:
  definition:
  - of or relating to the countryside as opposed to the city; living in or characteristic
    of farming or country life
  - living in or characteristic of farming or country life
  example:
  - rural people
  - large rural households
  - unpaved rural roads
  - an economy that is basically rural
  - rural electrification
  - rural free delivery
  ili: i11247
  members:
  - rural
  partOfSpeech: a
  similar:
  - 02058261-s
  - 02058442-s
  - 02058608-s
  - 02058929-s
  - 02059045-s
  - 02059217-s
  - 02059310-s
  - 02059434-s
  - 02059601-s
*/

type DictFile map[string]DictEntry

type DictEntry struct {
	Atributes    []string `yaml:"atribute" json:"atribute,omitempty"`
	Definitions  []string `yaml:"definition" json:"definition,omitempty"`
	Examples     any      `yaml:"example" json:"example,omitempty"`
	Members      []string `yaml:"members" json:"members,omitempty"`
	PartOfSpeech string   `yaml:"partOfSpeech" json:"part_of_speech,omitempty"`
	MeroPart     []string `yaml:"mero_part" json:"mero_part,omitempty"`
	DomainTopic  []string `yaml:"domain_topic" json:"domain_topic,omitempty"`
	Hypernym     []string `yaml:"hypernym" json:"hypernym,omitempty"`
	Similar      []string `yaml:"similar" json:"similar,omitempty"`

	Identifier string `yaml:"ili" json:"ili,omitempty"`
}

func (we *DictEntry) Words() []Word {
	var words = make([]Word, len(we.Members))

	for i, word := range we.Members {
		words[i] = Word{
			Word:         word,
			Definitions:  we.Definitions,
			PartOfSpeech: we.PartOfSpeech,
			Examples:     we.Examples,
		}
	}

	return words
}
