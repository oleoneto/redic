package repositories

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/oleoneto/redic/app/domain/external"

	"github.com/oleoneto/redic/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type DictionaryRepository struct {
	_db external.SqlEngineProtocol
}

// Explicit interface conformance
var _ external.WordRepositoryProtocol = (*DictionaryRepository)(nil)

func NewDictionaryRepository(database external.SqlEngineProtocol) *DictionaryRepository {
	return &DictionaryRepository{database}
}

// NewWords - Adds words to the dictionary database.
func (repo *DictionaryRepository) NewWords(ctx context.Context, words []external.NewWordInput) error {
	t, terr := repo._db.BeginTx(ctx, nil)
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

// AddWordDefinitions - Add new definitions to an existing word
func (repo *DictionaryRepository) AddWordDefinitions(ctx context.Context, data external.UpdateDefinitionInput) (external.AddDefinitionsOutput, error) {
	var res external.AddDefinitionsOutput

	query := `UPDATE words SET definitions = CONCAT(definitions, $3) WHERE word = $1 AND part_of_speech = $2 RETURNING id`

	_, err := repo._db.ExecContext(ctx, query, data.Word, data.PartOfSpeech, strings.Join(data.Definitions, "|"))
	if err != nil {
		return res, err
	}

	return res, nil
}

// GetWordDefinitions - Looks for the given word in the database dictionary and returns its definition(s).
func (repo *DictionaryRepository) GetWordDefinitions(ctx context.Context, data external.GetWordDefinitionsInput) (external.GetWordDefinitionsOutput, error) {
	var res external.GetWordDefinitionsOutput
	var args = []any{data.Word}

	partOfSpeechFilter := func() string {
		switch data.PartOfSpeech {
		case external.ALL:
			return `IS NOT NULL`
		}

		args = append(args, data.PartOfSpeech)
		return "= $2"
	}()

	query := fmt.Sprintf(`
	SELECT w.part_of_speech, d.definitions
	FROM
		definitions d
		JOIN words w ON 
			w.id = d.word_id
	WHERE
		w.word = $1
		AND w.part_of_speech %s
	ORDER BY
		w.word,
		w.part_of_speech
	`, partOfSpeechFilter)

	r, err := repo._db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	defer r.Close()

	res = external.GetWordDefinitionsOutput{Word: data.Word} //Term: data.Word}
	for r.Next() {
		var definition, partOfSpeech string
		if err := r.Scan(&partOfSpeech, &definition); err != nil {
			return res, err
		}

		res.Definitions = append(res.Definitions, struct {
			PartOfSpeech external.PartOfSpeech
			Definition   string
		}{
			PartOfSpeech: external.PartOfSpeech(partOfSpeech),
			Definition:   definition,
		})
	}

	return res, nil
}

// FindMatchingWords - Looks for all matching words for the provided word context.
func (repo *DictionaryRepository) GetDescribedWords(ctx context.Context, data external.GetDescribedWordsInput) (external.GetDescribedWordsOutput, error) {
	var res = external.GetDescribedWordsOutput{ProvidedDescriptions: data.Descriptions}

	var args = []any{}
	for _, d := range data.Descriptions {
		args = append(args, d)
	}

	matchers := helpers.EnumerateSQLArgs(
		len(args),
		0,
		func(index, counter int) string {
			if index > 0 {
				return fmt.Sprintf("OR MATCH $%d", index)
			}
			return fmt.Sprintf("MATCH $%d", index)
		},
	)

	query := fmt.Sprintf(`
	SELECT
		w.id,
		w.word,
		w.part_of_speech,
		redic_.definitions
	FROM
		words w
		JOIN redic_ ON 
			redic_.word_id = w.id
	WHERE
		redic_.definitions
		%s
	ORDER BY
		rank,
		w.word
	`, matchers)

	fmt.Println(query)

	os.Exit(1)

	r, err := repo._db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	defer r.Close()

	for r.Next() {
		var id, word, partOfSpeech, definition string

		if err := r.Scan(&id, &word, &partOfSpeech, &definition); err != nil {
			return res, err
		}

		res.MatchingWords = append(res.MatchingWords, struct {
			Id           string
			Word         string
			PartOfSpeech external.PartOfSpeech
			Definition   string
		}{
			Id:           id,
			Word:         word,
			PartOfSpeech: external.PartOfSpeech(partOfSpeech),
			Definition:   definition,
		})
	}

	return res, nil
}
