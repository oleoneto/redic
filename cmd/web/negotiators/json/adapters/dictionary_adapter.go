package adapters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oleoneto/redic/app/controllers"
	"github.com/oleoneto/redic/app/domain/types"
)

type DictionaryControllerAdapter struct {
	controller *controllers.DictionaryController
}

func NewDictionaryControllerAdapter(controller *controllers.DictionaryController) *DictionaryControllerAdapter {
	return &DictionaryControllerAdapter{controller: controller}
}

// ===========================================

func (ad *DictionaryControllerAdapter) GetWordDefinition(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	type queryParams struct {
		PartOfSpeech types.PartOfSpeech `query:"part_of_speech"`
		Verbatim     bool               `query:"verbatim"`
	}

	var q queryParams
	c.QueryParser(&q)

	var req = types.GetWordDefinitionsInput{
		Word:         c.Params("word"),
		PartOfSpeech: q.PartOfSpeech,
		Verbatim:     q.Verbatim,
	}

	res, err := ad.controller.GetDefinition(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(res)
}

func (ad *DictionaryControllerAdapter) FindWords(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	type queryParams struct {
		Query        string             `query:"q"`
		PartOfSpeech types.PartOfSpeech `query:"part_of_speech"`
		Cursor       string             `query:"cursor"`
	}

	var q queryParams
	c.QueryParser(&q)

	req := types.GetDescribedWordsInput{
		Tokens:       q.Query,
		PartOfSpeech: q.PartOfSpeech,
		Cursor:       q.Cursor,
	}

	res, err := ad.controller.FindMatchingWords(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(res)
}

// GET dictionary/words/:word
func (ad *DictionaryControllerAdapter) CreateWords(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req []types.NewWordInput

	b := c.Body()
	json.Unmarshal(b, &req)

	err := ad.controller.CreateWords(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

// PATCH dictionary/words/:word
func (ad *DictionaryControllerAdapter) UpdateWord(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req types.NewWordInput

	b := c.Body()
	json.Unmarshal(b, &req)

	err := ad.controller.CreateWords(ctx, []types.NewWordInput{req})
	if err != nil {
		return err
	}

	return c.JSON(req)
}
