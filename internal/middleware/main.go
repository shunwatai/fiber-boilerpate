package middleware

import (
	"golang-api-starter/internal/middleware/jwtcheck"
	"golang-api-starter/internal/middleware/logging"
	"golang-api-starter/internal/middleware/permissioncheck"

	"github.com/gofiber/fiber/v2"
)

type CustomMiddlewares struct{}

func (cm *CustomMiddlewares) CheckAccess(resourceName string) fiber.Handler {
	return (&permissioncheck.PermissionChecker{}).CheckAccess(resourceName)
}

func (cm *CustomMiddlewares) CheckJwt(ignorePaths ...string) fiber.Handler {
	return (&jwtcheck.JwtChecker{}).CheckJwt(ignorePaths...)
}

func (cm *CustomMiddlewares) Log() fiber.Handler {
	return (&logging.Logger{}).Log()
}

func NewCustMiddlewares() *CustomMiddlewares {
	return &CustomMiddlewares{}
}
