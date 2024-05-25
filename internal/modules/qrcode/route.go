package qrcode

import (
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var tableName = "qrcodes"
var Srvc = NewService()
var ctrl = NewController(Srvc)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares) {
	// r := router.Group("/qrcodes", jwtcheck.CheckFromHeader())
	r := router.Group("/api/qrcodes")
	r.Post("/from-pdf", GetQrcodeContentFromPdf)
}


// QrcodeFromPdf godoc
//
//	@Summary		Get Qrcode content from pdf
//	@Description	Get Qrcode content from pdf
//	@Tags			qrcodes
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"A document file like image or pdf"
//	@Router			/qrcodes [post]
func GetQrcodeContentFromPdf(c *fiber.Ctx) error {
	return ctrl.GetQrcodeContentFromPdf(c)
}

