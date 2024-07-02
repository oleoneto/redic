package repositories

import (
	"context"
	"fmt"
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

// GetWordExplanation - Looks for the given word in the database dictionary and returns its definition(s).
func (repo *DictionaryRepository) GetWordExplanation(ctx context.Context, data types.GetWordDefinitionsInput) (types.WordDefinitions, error) {
	var res = types.WordDefinitions{Definitions: []types.Definition{}, Word: data.Word}
	var args = []any{data.Word}

	partOfSpeechFilter := func() string {
		if data.PartOfSpeech == "" {
			return `IS NOT NULL`
		}

		args = append(args, data.PartOfSpeech)
		return "= $2"
	}()

	query := fmt.Sprintf(`
	SELECT
		d.id, d.word, d.part_of_speech, d.explanation, COALESCE(a.explicit, FALSE) explicit
	FROM
		dictionary d
		JOIN associations a
			ON a.explanation_id = d.explanation_id
			AND a.word_id = d.id
	WHERE
		d.word = $1
		AND d.part_of_speech %s
	`, partOfSpeechFilter)

	r, err := repo._db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	defer r.Close()

	for r.Next() {
		var id int
		var explicit bool
		var word, definition, partOfSpeech string
		if err := r.Scan(&id, &word, &partOfSpeech, &definition, &explicit); err != nil {
			return res, err
		}

		res.Definitions = append(res.Definitions, types.Definition{
			PartOfSpeech: types.PartOfSpeech(partOfSpeech),
			Definition:   definition,
			Explicit:     explicit,
		})
	}

	return res, nil
}

// SearchWords - Looks for all matching words for the provided word context.
func (repo *DictionaryRepository) SearchWords(ctx context.Context, data types.GetDescribedWordsInput) (types.WordMatches, error) {
	var res = types.WordMatches{ProvidedDescriptions: data.Tokens, MatchingWords: []types.MatchingWord{}}

	var args = []any{data.Tokens}

	filters := func() string {
		f := []string{}

		if data.Cursor != "" {
			f = append(f, fmt.Sprintf(`id > $%d`, len(args)+1))
			args = append(args, data.Cursor)
		}

		if data.PartOfSpeech != "" {
			args = append(args, data.PartOfSpeech)
			f = append(f, fmt.Sprintf(`part_of_speech = $%d`, len(args)+1))
		}

		if len(f) == 0 {
			return ""
		}

		return `WHERE ` + strings.Join(f, "AND")
	}()

	query := func() string {
		if data.Tokens != "" {
			return fmt.Sprintf(`
			SELECT
				word_id,
				word,
				w.part_of_speech,
				definition,
				highlight (redic_, 2, '<b>', '</b>') AS matched
			FROM
				redic_ ($1)
				JOIN words w ON redic_.word_id = w.id
			%s
			ORDER BY
				RANK
			LIMIT 100
			`, filters)
		}

		return fmt.Sprintf(`
		SELECT
			id,
			word,
			part_of_speech,
			explanation,
			"" AS highlight
		FROM
			dictionary
		%s
		ORDER BY
			word
		LIMIT 100
		`, filters)
	}()

	r, err := repo._db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	defer r.Close()

	for r.Next() {
		var id int
		var word, partOfSpeech, definition, highlight string

		if err := r.Scan(&id, &word, &partOfSpeech, &definition, &highlight); err != nil {
			return res, err
		}

		res.MatchingWords = append(res.MatchingWords, types.MatchingWord{
			Id:           id,
			Word:         word,
			PartOfSpeech: types.PartOfSpeech(partOfSpeech),
			Definition:   definition,
		})
	}

	if len(res.MatchingWords) > 0 {
		res.Cursor = fmt.Sprint(res.MatchingWords[len(res.MatchingWords)-1].Id)
	}

	return res, nil
}

func (repo *DictionaryRepository) IndexWords(ctx context.Context) error {
	query := `DELETE FROM redic_; INSERT INTO redic_ (word_id, word, definition) SELECT id, word, explanation FROM dictionary`

	if _, err := repo._db.ExecContext(ctx, query); err != nil {
		return err
	}

	return nil
}
