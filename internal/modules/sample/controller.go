package sample

import (
	"encoding/json"
	"fmt"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/notification/email"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	service *Service
}

func NewController(s *Service) *Controller {
	return &Controller{s}
}

var respCode = fiber.StatusInternalServerError

func (c *Controller) Ping(ctx *fiber.Ctx) error {
	return ctx.Status(200).JSON(fiber.Map{"data": "pong"})
}

func (c *Controller) SendEmail(ctx *fiber.Ctx) error {
	body := map[string]interface{}{}
	json.Unmarshal(ctx.BodyRaw(), &body)
	recipients := []string{}
	logger.Debugf("req body: %+v", body)

	for _, r := range body["to"].([]interface{}) {
		recipients = append(recipients, r.(string))
	}

	emailInfo := email.EmailInfo{
		To: recipients,
		MsgMeta: map[string]interface{}{
			"subject":          body["subject"].(string),
			"resetPasswordUrl": "https://yahoo.com",
		},
		MsgContent: body["message"].(string),
		Template:   template.Must(template.ParseGlob("web/template/reset-password/reset-email.gohtml")),
	}

	/* for send a simple text email */
	if err := email.SimpleEmail(emailInfo); err != nil {
		return ctx.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	/* for send a template email */
	if err := email.TemplateEmail(emailInfo); err != nil {
		return ctx.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.Status(200).JSON(fiber.Map{"data": body})
}

func (c *Controller) HalloPage(ctx *fiber.Ctx) error {
	tpl := template.Must(template.ParseFiles("web/template/hallo.gohtml", "web/template/base.gohtml"))

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	d := time.Now()
	dateStr := fmt.Sprintf("%02d/%s/%d %02d:%02d:%02d", d.Day(), d.Month().String(), d.Year(), d.Hour(), d.UTC().Minute(), d.UTC().Second())
	envs := map[string]interface{}{
		"env":       cfg.ServerConf.Env,
		"db driver": cfg.DbConf.Driver,
	}

	// err := tpl.Execute(fctx.Fctx.Response().BodyWriter(), map[string]interface{}{"date": dateStr, "envs": envs})
	err := tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml",map[string]interface{}{"date": dateStr, "envs": envs})

	logger.Debugf("tpl exe err: %+v", err)
	return err
}
