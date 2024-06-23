package adapters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oleoneto/redic/app/controllers"
	"github.com/oleoneto/redic/app/domain/external"
)

type DictionaryControllerAdapter struct {
	controller *controllers.DictionaryController
}

func NewDictionaryControllerAdapter(controller *controllers.DictionaryController) *DictionaryControllerAdapter {
	return &DictionaryControllerAdapter{controller: controller}
}

func (ad *DictionaryControllerAdapter) CreateWords(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req []external.NewWordInput

	b := c.Body()
	json.Unmarshal(b, &req)

	err := ad.controller.CreateWords(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

func (ad *DictionaryControllerAdapter) UpdateWord(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req external.UpdateDefinitionInput

	b := c.Body()
	json.Unmarshal(b, &req)

	res, err := ad.controller.UpdateDefinition(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(res)
}

func (ad *DictionaryControllerAdapter) GetWordDefinition(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req external.GetWordDefinitionsInput

	b := c.Body()
	json.Unmarshal(b, &req)

	res, err := ad.controller.GetDefinition(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(res)
}

func (ad *DictionaryControllerAdapter) FindWords(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	var req external.GetDescribedWordsInput

	b := c.Body()
	json.Unmarshal(b, &req)

	res, err := ad.controller.FindMatchingWords(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(res)
}
