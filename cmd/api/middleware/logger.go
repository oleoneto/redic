package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RequestLoggerMiddleware(c *fiber.Ctx) error {
	s := time.Now()

	err := c.Next()

	logrus.Infoln(
		c.Response().StatusCode(),
		c.Protocol(),
		c.Method(),
		c.OriginalURL(),
		c.Route().Name,
		time.Since(s),
		err,
	)

	return err
}
