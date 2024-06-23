package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/cmd/api/adapters"
)

func CreateAPI(config fiber.Config) *fiber.App {
	server := fiber.New(config)

	// MARK: Middleware

	server.Use(recover.New(recover.Config{EnableStackTrace: true}))
	server.Use(requestid.New(requestid.ConfigDefault))
	// server.Use()

	// MARK: Routes

	var dictionaryAdapter = adapters.NewDictionaryControllerAdapter(&app.DictionaryController)

	server.Route("/dictionary", func(router fiber.Router) {
		router.Post("/words", dictionaryAdapter.CreateWords).Name("create-words")
		router.Get("/words/search", dictionaryAdapter.FindWords).Name("find-word")
		router.Get("/words/:word", dictionaryAdapter.GetWordDefinition).Name("get-word-definition")
		router.Patch("/words/:word", dictionaryAdapter.UpdateWord).Name("update-word-definition")
	})

	return server
}
