package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/oleoneto/redic/db"
	"github.com/oleoneto/redic/pkg/core"
	"github.com/oleoneto/redic/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type Query struct{ database db.SqlEngineProtocol }

type (
	Definition struct {
		Word, PartOfSpeech, Content string
	}

	Word struct {
		Word, PartOfSpeech string
		Definitions        []string
	}

	SearchResult struct {
		Words []Word `yaml:"words" json:"words"`
	}

	DefinitionResult struct {
		Term        string
		Definitions []Definition `yaml:"definitions" json:"definitions"`
	}
)

func NewQuery(database db.SqlEngineProtocol) *Query { return &Query{database} }

func (Q *Query) Define(ctx context.Context, term string, partOfSpeech string) (DefinitionResult, error) {
	p := func() string {
		if partOfSpeech == "" {
			return ""
		}
		return fmt.Sprintf("AND words.part_of_speech = '%s'", partOfSpeech)
	}()

	q := fmt.Sprintf(`
	SELECT
		words.part_of_speech,
		d.definitions
	FROM
		definitions d
		JOIN words ON words.id = d.word_id
	WHERE
		words.word = $1
	%s
	ORDER BY
		words.word,
		words.part_of_speech
	`, p)

	r, err := Q.database.QueryContext(ctx, q, term)
	if err != nil {
		return DefinitionResult{}, err
	}
	defer r.Close()

	var results = DefinitionResult{Term: term}
	for r.Next() {
		var definition = Definition{Word: term}
		if err := r.Scan(&definition.PartOfSpeech, &definition.Content); err != nil {
			return results, err
		}

		results.Definitions = append(results.Definitions, definition)
	}

	return results, nil
}

func (Q *Query) Search(ctx context.Context, terms ...string) (SearchResult, error) {
	q := `
	SELECT
		words.word,
		words.part_of_speech,
		redic_.definitions
	FROM
		words
		JOIN redic_ ON redic_.word_id = words.id
	WHERE
		redic_.definitions MATCH $1
	ORDER BY
		rank,
		words.word
	`

	r, err := Q.database.QueryContext(ctx, q, strings.Join(terms, " "))
	if err != nil {
		return SearchResult{}, err
	}
	defer r.Close()

	var results = SearchResult{}

	for r.Next() {
		var w Word
		var d string

		if err := r.Scan(&w.Word, &w.PartOfSpeech, &d); err != nil {
			return results, err
		}

		w.Definitions = append(w.Definitions, strings.Split(d, "|")...)
		results.Words = append(results.Words, w)
	}

	return results, nil
}

func (Q *Query) SaveWords(ctx context.Context, words []core.Word) error {
	t, terr := Q.database.BeginTx(ctx, nil)
	if terr != nil {
		return terr
	}

	var values, definitionValues string

	var args []any
	var definitionArgs []any

	for _, word := range words {
		values += `(?, ?, ?), `        // word, part_of_speech, ili
		definitionValues += `(?, ?), ` // word_id, definition
		args = append(args, word.Word, word.PartOfSpeech, word.EntryCode)
	}

	values = strings.TrimSpace(values)
	values = strings.TrimRight(values, ",")

	definitionValues = strings.TrimSpace(definitionValues)
	definitionValues = strings.TrimRight(definitionValues, ",")

	query := fmt.Sprintf(`
	INSERT INTO
		words (word, part_of_speech, ili)
	VALUES %s
		ON CONFLICT (word, part_of_speech, ili) DO NOTHING
	RETURNING id`, values)

	r, err := t.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return err
	}
	defer r.Close()

	var wordIds []int
	for r.Next() {
		var id int
		err := r.Scan(&id)
		if err != nil {
			t.Rollback()
			return err
		}

		wordIds = append(wordIds, id)
		definitionArgs = append(definitionArgs, id, strings.Join(words[len(wordIds)-1].Definitions, "| "))
	}

	tids := len(wordIds)
	twords := len(words)

	if tids != twords {
		t.Rollback()
		return nil // fmt.Errorf(`received %d saved ids for %d words`, tids, twords)
	}

	dquery := fmt.Sprintf(`
	INSERT INTO
		definitions (word_id, definitions)
	VALUES %s
	`, definitionValues)

	_, err = t.ExecContext(ctx, dquery, definitionArgs...)
	if err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return err
	}

	return t.Commit()
}
