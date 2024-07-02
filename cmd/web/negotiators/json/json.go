package json

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/cmd/web/negotiators/json/adapters"
)

// Router - decorates the provided application with API-only routes.
func Router(router fiber.Router) {
	var dictionaryAdapter = adapters.NewDictionaryControllerAdapter(&app.DictionaryController)

	// i.e /words/alone?part_of_speech=n
	router.Get("/words/:word", dictionaryAdapter.GetWordDefinition).Name("get-word-definition")

	// i.e /dictionary/words?q=present_location&part_of_speech=n
	router.Get("/words", dictionaryAdapter.FindWords).Name("find-word")

	// router.Post("/words", dictionaryAdapter.CreateWords).Name("create-words")
	// router.Patch("/words/:word", dictionaryAdapter.UpdateWord).Name("update-word-definition")
}
