package interfaces

import "github.com/gofiber/fiber/v2"

type ICustomMiddlewares interface {
	CheckAccess(string) fiber.Handler
	CheckJwt() fiber.Handler
	Log() fiber.Handler
}
