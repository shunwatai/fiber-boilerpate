package web

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/middleware/jwtcheck"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg  = config.Cfg
	Srvc = &Service{}
	ctrl = &Controller{}
)

func GetRoutes(router fiber.Router) {
	Srvc = NewService()
	ctrl = NewController(Srvc)

	router.Get("/home", jwtcheck.CheckJwt(), ctrl.HomePage)
	router.Get("/error", ctrl.ErrorPage)
}
