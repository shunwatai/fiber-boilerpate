package helper

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

type HttpErr struct {
	Code int
	Err  error
}

func (h *HttpErr) Error() string {
	return h.Err.Error()
}

func (ctx *FiberCtx) JsonResponse(respCode int, data map[string]interface{}) error {
	return ctx.Fctx.
		Status(respCode).
		JSON(data)
}

func (ctx *FiberCtx) ErrResponse(respCode int, err error) error {
	reqHeader := ctx.Fctx.GetReqHeaders()
	isHtml := strings.Contains(reqHeader["Accept"][0], "text/html")

	if isHtml {
		ctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return ctx.Fctx.
			Status(respCode).
			Redirect("/error", fiber.StatusPermanentRedirect)
	}

	return ctx.Fctx.
		Status(respCode).
		JSON(map[string]interface{}{"message": err.Error()})
}
