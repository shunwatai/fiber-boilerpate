package oauth

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg  = config.Cfg
	Srvc = &Service{}
	ctrl = &Controller{}
)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares) {
	Srvc = NewService()
	ctrl = NewController(Srvc)

	router.Get("/sign-in", OAuthProviderPage)

	// oauth for google
	r := router.Group("/api/oauth")
	r.Get("/:provider/callback", OAuthGetAuth)
	r.Get("/:provider/login", OAuthLogin)
	r.Get("/logout", OAuthLogout)
}

func OAuthGetAuth(c *fiber.Ctx) error {
	return ctrl.OAuthGetAuth(c)
}
func OAuthLogin(c *fiber.Ctx) error {
	return ctrl.OAuthLogin(c)
}
func OAuthLogout(c *fiber.Ctx) error {
	return ctrl.OAuthLogout(c)
}
func OAuthProviderPage(c *fiber.Ctx) error {
	return ctrl.OAuthProviderPage(c)
}
