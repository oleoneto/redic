package repositories

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/oleoneto/redic/app/domain/protocols"
	"github.com/oleoneto/redic/app/domain/types"

	"github.com/oleoneto/redic/app/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type DictionaryRepository struct {
	_db protocols.SqlBackend
}

// Explicit interface conformance check
var _ protocols.DictionaryBackend = (*DictionaryRepository)(nil)

func NewDictionaryRepository(database protocols.SqlBackend) *DictionaryRepository {
	return &DictionaryRepository{database}
}

// NewWords - Adds words to the dictionary database.
func (repo *DictionaryRepository) NewWords(ctx context.Context, words []types.NewWordInput) error {
	t := repo._db

	/* Creates a new word entry if one does not yet exist */
	newWord := `
	INSERT INTO words(text, part_of_speech)
		VALUES($1, $2)
		ON CONFLICT (text, part_of_speech)
		DO UPDATE SET text = $1, part_of_speech = $2
	RETURNING id
	`

	/* Creates a new explanation if one does not yet exist */
	newExplanation := `
	INSERT INTO explanations(text)
		VALUES($1)
		ON CONFLICT(text)
		DO UPDATE SET text = $1
	RETURNING id
	`

	newAssociation := `
		INSERT INTO associations(word_id, explanation_id) VALUES($1, $2) ON CONFLICT(word_id, explanation_id) DO NOTHING
	`

	for _, item := range words {
		var wordId, explanationId int64

		if item.Word == "time" {
			fmt.Println(item)
		}

		if r := t.QueryRowContext(ctx, newWord, item.Word, item.PartOfSpeech); r != nil {
			if err := r.Scan(&wordId); err != nil {
				logrus.Errorln("failed to add word", item.Word, "part_of_speech", item.PartOfSpeech)
				return err
			}
		}

		if e := t.QueryRowContext(ctx, newExplanation, item.Definition); e != nil {
			if err := e.Scan(&explanationId); err != nil {
				logrus.Errorln("failed to add explanation", item.Definition)
				return err
			}
		}

		_, err := t.ExecContext(ctx, newAssociation, wordId, explanationId)
		if err != nil {
			logrus.Errorln("failed to associate id and explanation for", item.Word, wordId, explanationId)
			return err
		}
	}

	return nil
}

// AddWordDefinitions - Add new definitions to an existing word
func (repo *DictionaryRepository) AddWordDefinitions(ctx context.Context, data types.UpdateDefinitionInput) (types.Definitions, error) {
	var res types.Definitions

	t, terr := repo._db.BeginTx(ctx, nil)
	if terr != nil {
		return res, terr
	}

	values := `(?, ?, ?), `
	args := []any{data.Word, data.PartOfSpeech, data.Definitions}

	values = strings.TrimSpace(values)
	values = strings.TrimRight(values, ",")

	query := fmt.Sprintf(`
	INSERT INTO
		dictionary (word, part_of_speech, explanation)
	VALUES %s
	RETURNING *`, values)

	r, err := t.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return res, err
	}
	defer r.Close()

	return res, t.Commit()
}

// GetWordDefinitions - Looks for the given word in the database dictionary and returns its definition(s).
func (repo *DictionaryRepository) GetWordDefinitions(ctx context.Context, data types.GetWordDefinitionsInput) (types.WordDefinitions, error) {
	var res types.WordDefinitions
	var args = []any{data.Word}

	partOfSpeechFilter := func() string {
		if data.PartOfSpeech == "" {
			return `IS NOT NULL`
		}

		args = append(args, data.PartOfSpeech)
		return "= $2"
	}()

	query := fmt.Sprintf(`SELECT id, word, part_of_speech, explanation FROM dictionary WHERE word = $1 AND part_of_speech %s`, partOfSpeechFilter)

	r, err := repo._db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	defer r.Close()

	res = types.WordDefinitions{Word: data.Word} //Term: data.Word}
	for r.Next() {
		var id int
		var word, definition, partOfSpeech string
		if err := r.Scan(&id, &word, &partOfSpeech, &definition); err != nil {
			return res, err
		}

		res.Definitions = append(res.Definitions, struct {
			PartOfSpeech types.PartOfSpeech `json:"part_of_speech"`
			Definition   string             `json:"definition"`
		}{
			PartOfSpeech: types.PartOfSpeech(partOfSpeech),
			Definition:   definition,
		})
	}

	return res, nil
}

// FindMatchingWords - Looks for all matching words for the provided word context.
func (repo *DictionaryRepository) GetDescribedWords(ctx context.Context, data types.GetDescribedWordsInput) (types.DescribedWords, error) {
	var res = types.DescribedWords{ProvidedDescriptions: data.Descriptions}

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
			PartOfSpeech types.PartOfSpeech
			Definition   string
		}{
			Id:           id,
			Word:         word,
			PartOfSpeech: types.PartOfSpeech(partOfSpeech),
			Definition:   definition,
		})
	}

	return res, nil
}

func (repo *DictionaryRepository) IndexWords(ctx context.Context) error {
	query := `INSERT INTO redic_ (word_id, word, definition) SELECT id, word, explanation FROM dictionary`

	if _, err := repo._db.ExecContext(ctx, query); err != nil {
		return err
	}

	return nil
}
