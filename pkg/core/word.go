package core

import (
	"context"
	"strings"

	"github.com/oleoneto/redic/db"
	"github.com/oleoneto/redic/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type Word struct {
	PartOfSpeech string   // a
	Word         string   // emerging
	Definitions  []string // comming into existence
	Examples     any      // an emergent republic
}

// Save a word by creating an entry for it. Each definition should link to the word_id.
func (w *Word) Save(ctx context.Context, db db.SqlEngineProtocol) error {
	t, terr := db.BeginTx(ctx, nil)
	if terr != nil {
		return terr
	}

	query := `
		INSERT INTO
			words (word, part_of_speech)
		VALUES ($1, $2)
		ON CONFLICT
			(word, part_of_speech)
		DO NOTHING
		RETURNING
			id
		`
	r, err := t.QueryContext(ctx, query, w.Word, w.PartOfSpeech)
	if err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return err
	}
	defer r.Close()

	if !r.Next() {
		t.Rollback()
		return nil
	}

	var id int64
	if err := r.Scan(&id); err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return err
	}

	// MARK: Persist definitions

	query = `
	INSERT INTO
	definitions (word_id, definitions)
	VALUES ($1, $2)
	RETURNING *
	`

	definitions := strings.Join(w.Definitions, "|")

	_, err = t.ExecContext(
		ctx,
		query,
		id,
		definitions,
	)
	if err != nil {
		logrus.Errorln(helpers.GetCurrentFuncName(), err)
		t.Rollback()
		return err
	}

	// TODO: Update redic_ table

	t.Commit()

	return nil
}
