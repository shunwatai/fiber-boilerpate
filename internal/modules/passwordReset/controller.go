package passwordReset

import (
	"errors"
	"fmt"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/user"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type Controller struct {
	service *Service
}

func sanitise(prs PasswordResets) {
	for _, pr := range prs {
		pr.TokenHash = nil
	}
}

func NewController(s *Service) *Controller {
	return &Controller{s}
}

var respCode = fiber.StatusInternalServerError

func (c *Controller) Get(ctx *fiber.Ctx) error {
	logger.Debugf("passwordReset ctrl\n")
	fctx := &helper.FiberCtx{Fctx: ctx}
	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	results, pagination := c.service.Get(paramsMap)

	respCode = fiber.StatusOK
	return fctx.JsonResponse(
		respCode,
		map[string]interface{}{"data": results, "pagination": pagination},
	)
}

func (c *Controller) GetById(ctx *fiber.Ctx) error {
	logger.Debugf("passwordReset ctrl\n")
	fctx := &helper.FiberCtx{Fctx: ctx}
	id := fctx.Fctx.Params("id")
	paramsMap := map[string]interface{}{"id": id}
	results, err := c.service.GetById(paramsMap)

	if err != nil {
		respCode = fiber.StatusNotFound
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"message": err.Error()},
		)
	}

	respCode = fiber.StatusOK
	return fctx.JsonResponse(respCode, map[string]interface{}{"data": results[0]})
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	logger.Debugf("passwordReset ctrl create\n")
	c.service.ctx = ctx
	passwordReset := &PasswordReset{}
	passwordResets := []*PasswordReset{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	passwordResetErr, parseErr := reqCtx.Payload.ParseJsonToStruct(passwordReset, &passwordResets)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if passwordResetErr == nil {
		passwordResets = append(passwordResets, passwordReset)
	}
	// logger.Debugf("passwordResetErr: %+v, passwordResetsErr: %+v\n", passwordResetErr, passwordResetsErr)
	// for _, t := range passwordResets {
	// 	logger.Debugf("passwordResets: %+v\n", t)
	// }

	for _, passwordReset := range passwordResets {
		if validErr := helper.ValidateStruct(*passwordReset); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}

		if passwordReset.Id == nil {
			continue
		} else if existing, err := c.service.GetById(map[string]interface{}{
			"id": passwordReset.GetId(),
		}); err == nil && passwordReset.CreatedAt == nil {
			passwordReset.CreatedAt = existing[0].CreatedAt
		}
		// logger.Debugf("passwordReset? %+v\n", passwordReset)
	}

	// return []*PasswordReset{}
	results, httpErr := c.service.Create(passwordResets)
	sanitise(results)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusCreated
	if passwordResetErr == nil && len(results) > 0 {
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"data": results[0]},
		)
	}
	return fctx.JsonResponse(
		respCode,
		map[string]interface{}{"data": results},
	)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	logger.Debugf("passwordReset ctrl update\n")

	passwordReset := &PasswordReset{}
	passwordResets := []*PasswordReset{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	passwordResetErr, parseErr := reqCtx.Payload.ParseJsonToStruct(passwordReset, &passwordResets)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if passwordResetErr == nil {
		passwordResets = append(passwordResets, passwordReset)
	}

	for _, passwordReset := range passwordResets {
		if validErr := helper.ValidateStruct(*passwordReset); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if passwordReset.Id == nil && passwordReset.MongoId == nil {
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{"message": "please ensure all records with id for PATCH"},
			)
		}

		conditions := map[string]interface{}{}
		conditions["id"] = passwordReset.GetId()

		existing, err := c.service.GetById(conditions)
		if len(existing) == 0 {
			respCode = fiber.StatusNotFound
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{
					"message": errors.Join(
						errors.New("cannot update non-existing records..."),
						err,
					).Error(),
				},
			)
		} else if passwordReset.CreatedAt == nil {
			passwordReset.CreatedAt = existing[0].CreatedAt
		}
	}

	results, httpErr := c.service.Update(passwordResets)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusOK
	if passwordResetErr == nil && len(results) > 0 {
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"data": results[0]},
		)
	}
	return fctx.JsonResponse(
		respCode,
		map[string]interface{}{"data": results},
	)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	logger.Debugf("passwordReset ctrl delete\n")
	// body := map[string]interface{}{}
	// json.Unmarshal(c.BodyRaw(), &body)
	// logger.Debugf("req body: %+v\n", body)
	delIds := struct {
		Ids []int64 `json:"ids" validate:"required,unique"`
	}{}

	mongoDelIds := struct {
		Ids []string `json:"ids" validate:"required,unique"`
	}{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	intIdsErr, strIdsErr := reqCtx.Payload.ParseJsonToStruct(&delIds, &mongoDelIds)
	if intIdsErr != nil && strIdsErr != nil {
		logger.Errorf("failed to parse req json, %+v\n", errors.Join(intIdsErr, strIdsErr).Error())
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": errors.Join(intIdsErr, strIdsErr).Error()})
	}
	if len(delIds.Ids) == 0 && len(mongoDelIds.Ids) == 0 {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": "please check the req json like the follow: {\"ids\":[]}"})
	}
	logger.Debugf("deletedIds: %+v, mongoIds: %+v\n", delIds, mongoDelIds)

	var (
		results []*PasswordReset
		err     error
	)

	if cfg.DbConf.Driver == "mongodb" {
		results, err = c.service.Delete(mongoDelIds.Ids)
	} else {
		idsString, _ := helper.ConvertNumberSliceToString(delIds.Ids)
		results, err = c.service.Delete(idsString)
	}

	if err != nil {
		logger.Errorf("failed to delete, err: %+v\n", err.Error())
		respCode = fiber.StatusNotFound
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"message": err.Error()},
		)
	}

	respCode = fiber.StatusOK
	return fctx.JsonResponse(
		respCode,
		map[string]interface{}{"data": results},
	)
}

func SetResetTokenInCookie(result map[string]interface{}, c *fiber.Ctx) {
	env := cfg.ServerConf.Env
	resetToken := result["accessToken"].(string)
	cookie := &fiber.Cookie{
		Name:     "accessToken",
		Value:    resetToken,
		Expires:  time.Now().Add(time.Minute * 10), // 10 mins
		HTTPOnly: true,
		Secure:   env == "prod",
		Path:     "/",
	}

	c.Cookie(cookie)
	delete(result, "accessToken")
}

func (c *Controller) SendResetEmailPage(ctx *fiber.Ctx) error {
	// data for template
	data := map[string]interface{}{
		"errMessage": nil,
	}

	tpl := template.Must(template.ParseGlob("web/template/reset-password/send-reset-email.gohtml"))

	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) SendResetEmail(ctx *fiber.Ctx) error {
	// data for template
	data := map[string]interface{}{
		"errMessage": nil,
	}

	html := `
	<div>{{$.errMessage}}</div>
	<div>{{$.message}}</div>
	`
	tpl, _ := template.New("change-password").Parse(html)

	fctx := &helper.FiberCtx{Fctx: ctx}

	u := new(user.User)
	if err := fctx.Fctx.BodyParser(u); err != nil {
		logger.Errorf("BodyParser err: %+v", err)
		data["errMessage"] = "something went wrong: failed to parse request json"
	}

	_, httpErr := c.service.Create(PasswordResets{&PasswordReset{Email: *u.Email}})
	if httpErr.Err != nil {
		logger.Errorf("BodyParser err: %+v", httpErr.Err)
		data["errMessage"] = "email doesn't match with any existing users..."
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	data["message"] = "Reset email has been sent, please check your mailbox"
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) PasswordResetPage(ctx *fiber.Ctx) error {
	// data for template
	data := map[string]interface{}{
		"errMessage": nil,
	}
	tpl := template.Must(template.ParseGlob("web/template/reset-password/reset-form.gohtml"))

	fctx := &helper.FiberCtx{Fctx: ctx}
	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	logger.Debugf("paramsMap: %v", paramsMap["token"])
	token := paramsMap["token"].(string)
	email := paramsMap["email"].(string)
	users, _ := user.Srvc.Get(map[string]interface{}{"email": email, "exactMatch": map[string]bool{"email": true}})
	logger.Debugf("users: %v", len(users))
	if len(users) == 0 {
		data["errMessage"] = fmt.Sprintf("email not found: %s", email)
		fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	passwordResets, _ := c.service.Get(map[string]interface{}{"user_id": users[0].GetId(), "is_used": false})
	if len(passwordResets) == 0 {
		data["errMessage"] = "something went wrong, please try to send reset password again"
		fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	err := bcrypt.CompareHashAndPassword([]byte(*passwordResets[0].TokenHash), []byte(token))
	if err != nil {
		data["errMessage"] = "something went wrong, please try to send reset password again"
		fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	resetToken, err := GetResetJwtToken(passwordResets[0])
	logger.Debugf("resetToken: %v", resetToken)
	if err != nil {
		logger.Errorf("GetResetJwtToken err: %+v", err)
	}

	respCode = fiber.StatusOK

	data["token"] = token
	data["userId"] = users[0].Id
	data["name"] = users[0].Name
	SetResetTokenInCookie(map[string]interface{}{"accessToken": resetToken}, fctx.Fctx)
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) ChangePassword(ctx *fiber.Ctx) error {
	data := fiber.Map{}
	html := `
	<div>{{$.message}}</div>
	<div>go to login page</div>
	`
	tpl, _ := template.New("change-password").Parse(html)

	u := new(user.User)
	fctx := &helper.FiberCtx{Fctx: ctx}

	if err := fctx.Fctx.BodyParser(u); err != nil {
		logger.Errorf("BodyParser err: %+v", err)
		data["message"] = "something went wrong: failed to parse request json"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	users, httpErr := user.Srvc.Update(user.Users{u})
	if httpErr.Err != nil {
		logger.Errorf("user Update err: %+v", httpErr.Err.Error())
		data["message"] = "something went wrong: failed to reset password"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	data["message"] = fmt.Sprintf("reset success for %s", users[0].Name)
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}
