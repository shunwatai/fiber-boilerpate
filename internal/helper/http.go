package helper

import (
	"html/template"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
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

var accessDeniedCodes = []int{fiber.StatusUnauthorized, fiber.StatusForbidden}

func (ctx *FiberCtx) ErrResponse(respCode int, err error) error {
	reqHeader := ctx.Fctx.GetReqHeaders()
	isHtml := strings.Contains(reqHeader["Accept"][0], "text/html")
	isHxReq := reqHeader["Hx-Request"] != nil && reqHeader["Hx-Request"][0] == "true"

	if isHtml {
		return ctx.handleHtmlError(respCode)
	}

	if isHxReq && slices.Contains(accessDeniedCodes, respCode) {
		return ctx.handleHxUnauthorizedError(respCode, errors.New("Insufficient permission"))
	}

	return ctx.Fctx.Status(respCode).JSON(map[string]interface{}{"message": err.Error()})
}

// handleHtmlError redirect to error page
func (ctx *FiberCtx) handleHtmlError(respCode int) error {
	ctx.Fctx.Set("Expires", "Tue, 03 Jul 2001 06:00:00 GMT")
	ctx.Fctx.Set("Last-Modified", "{now} GMT")
	ctx.Fctx.Set("Cache-Control", "max-age=0, no-cache, private, must-revalidate, proxy-revalidate")
	if slices.Contains(accessDeniedCodes, respCode) {
		return ctx.Fctx.Status(respCode).Redirect("/unauthorised", fiber.StatusTemporaryRedirect)
	}
	return ctx.Fctx.Status(respCode).Redirect("/error", fiber.StatusTemporaryRedirect)
}

// handleHxUnauthorizedError responses html of popup error message
func (ctx *FiberCtx) handleHxUnauthorizedError(respCode int, err error) error {
	ctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup". }}`
	tpl, _ = tpl.New("message").Parse(html)
	return tpl.Execute(ctx.Fctx.Status(respCode).Response().BodyWriter(), fiber.Map{"errMessage": err.Error()})
}
