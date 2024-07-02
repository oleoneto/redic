package html

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Router - decorates the provided server with HTML-only routes.
func Router(router fiber.Router) {
	router.
		Get("", func(c *fiber.Ctx) error {
			return c.Render(
				"templates/index",
				fiber.Map{
					"time": time.Now(),
				},
				"templates/layouts/base",
			)
		}).
		Name("web:index")

	router.
		Get("/:letter", func(c *fiber.Ctx) error {
			return c.Render(
				"templates/letter",
				fiber.Map{
					"time":   time.Now(),
					"title":  "",
					"letter": c.Params("letter"),
				},
				"templates/layouts/base",
			)
		}).
		Name("web:letter")

	router.
		Get("/words/:word", func(c *fiber.Ctx) error {
			return c.Render(
				"templates/word",
				fiber.Map{
					"time":  time.Now(),
					"title": "",
					"word":  c.Params("word"),
				},
				"templates/layouts/base",
			)
		}).
		Name("web:word")
}
