package helper

import (
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v2"
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
	isHxReq := false
	if reqHeader["Hx-Request"] != nil && reqHeader["Hx-Request"][0] == "true" {
		isHxReq = true
	}

	if isHtml {
		if respCode == fiber.StatusUnauthorized {
			ctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return ctx.Fctx.
				Status(respCode).
				Redirect("/unauthorised", fiber.StatusPermanentRedirect)
		} else {
			return ctx.Fctx.
				Status(respCode).
				Redirect("/error", fiber.StatusPermanentRedirect)
		}
	}

	if isHxReq && respCode == fiber.StatusUnauthorized {
		ctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		tmplFiles := []string{"web/template/parts/popup.gohtml"}
		tpl := template.Must(template.ParseFiles(tmplFiles...))

		html := `{{ template "popup" . }}`
		tpl, _ = tpl.New("message").Parse(html)
		return tpl.Execute(ctx.Fctx.Status(respCode).Response().BodyWriter(), fiber.Map{"errMessage": "Insufficient permission"})
	}

	return ctx.Fctx.
		Status(respCode).
		JSON(map[string]interface{}{"message": err.Error()})
}
