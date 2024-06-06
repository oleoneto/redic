package cli

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/oleoneto/redic/pkg/query"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Args:    cobra.ArbitraryArgs,
	Short:   "Search for words matching a definition.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()

		q := query.NewQuery(state.Database)

		words, err := q.Search(ctx, args...)
		if err != nil {
			panic(err)
		}

		state.Writer.Print(QuerySearchResult(words))
	},
}

// -------------------------------------------

type QueryWord query.Word
type QueryDefinition query.Definition
type QuerySearchResult query.SearchResult
type QueryDefinitionResult query.DefinitionResult

func (w QueryWord) String() string {
	/**
	water (n)
		a liquid necessary for the life of most animals and plants
	*/

	c := "{{- range .Definitions }}{{ .Word }} ({{ .PartOfSpeech }})\n\t{{ . }}\n{{- end }}"
	t, err := template.New("word").Parse(c)
	if err != nil {
		panic(err)
	}

	var bf bytes.Buffer
	if err := t.Execute(&bf, w); err != nil {
		panic(err)
	}

	return bf.String()
}

func (d QueryDefinition) String() string {
	/**
	water (n)
		a liquid necessary for the life of most animals and plants
	*/

	c := "{{ .Word }} ({{ .PartOfSpeech }})\n  {{ .Content }}\n"
	t, err := template.New("definition").Parse(c)
	if err != nil {
		panic(err)
	}

	var bf bytes.Buffer
	if err := t.Execute(&bf, d); err != nil {
		panic(err)
	}

	return bf.String()
}

func (sr QuerySearchResult) String() string {
	/**
	here (n)
		the present location; this place; location, proximal pronoun; demonstrative pronoun, location; quantifier: demonstrative determiner, singular, proximal

	endemic	(n)
		a disease that is constantly present to a greater or lesser degree in people of a certain class or in people living in a particular location

	endemic disease	(n)
		a disease that is constantly present to a greater or lesser degree in people of a certain class or in people living in a particular location
	*/

	c := `{{ range .Words }}{{ template "word" . }}{{ end }}`
	t, err := template.New("search-results").Parse(c)
	if err != nil {
		panic(err)
	}

	var bf bytes.Buffer
	if err := t.Execute(&bf, sr); err != nil {
		panic(err)
	}

	return bf.String()
}

func (dr QueryDefinitionResult) String() string {
	c := "{{ range .Definitions }}{{ . }}\n{{ end }}"
	t, err := template.New("definition-results").Parse(c)
	if err != nil {
		panic(err)
	}

	var bf bytes.Buffer
	if err := t.Execute(&bf, dr); err != nil {
		panic(err)
	}

	return bf.String()
}
