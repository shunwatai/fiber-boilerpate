package interfaces

import "github.com/gofiber/fiber/v2"

type ICustomMiddlewares interface {
	CheckAccess(string) fiber.Handler
	CheckJwt(...string) fiber.Handler
	Log() fiber.Handler
}
