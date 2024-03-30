package user

import (
	"errors"
	"fmt"
	"golang-api-starter/internal/auth"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"html/template"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	service *Service
}

func sanitise(users Users) {
	for _, u := range users {
		u.Password = nil
	}
}

func NewController(s *Service) *Controller {
	return &Controller{s}
}

var mu sync.Mutex
var respCode = fiber.StatusInternalServerError

/* SetTokensInCookie is a helper for Login & Refresh funcs for setting the cookies in response */
func SetTokensInCookie(result map[string]interface{}, c *fiber.Ctx) error {
	if result["refreshToken"] == nil && result["accessToken"] == nil {
		return logger.Errorf("missing required 'accessToken' & 'refreshToken'")
	}
	env := cfg.ServerConf.Env
	refreshToken := result["refreshToken"].(string)
	cookie := &fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Hour * 720), // 30 days
		HTTPOnly: true,
		Secure:   env == "prod",
		Path:     "/",
	}
	c.Cookie(cookie)

	accessToken := result["accessToken"].(string)
	cookie = &fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(time.Hour * 720), // 30 days
		HTTPOnly: true,
		Secure:   env == "prod",
		Path:     "/",
	}
	c.Cookie(cookie)

	delete(result, "refreshToken")
	return nil
}

func (c *Controller) Get(ctx *fiber.Ctx) error {
	logger.Debugf("user ctrl\n")
	fctx := &helper.FiberCtx{Fctx: ctx}
	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	results, pagination := c.service.Get(paramsMap)
	sanitise(results)

	respCode = fiber.StatusOK
	return fctx.JsonResponse(
		respCode,
		map[string]interface{}{"data": results, "pagination": pagination},
	)
}

func (c *Controller) GetById(ctx *fiber.Ctx) error {
	logger.Debugf("user ctrl\n")
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
	logger.Debugf("user ctrl create\n")
	c.service.ctx = ctx
	user := &User{}
	users := []*User{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	userErr, parseErr := reqCtx.Payload.ParseJsonToStruct(user, &users)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if userErr == nil {
		users = append(users, user)
	}

	for _, user := range users {
		if validErr := helper.ValidateStruct(*user); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if user.Id == nil {
			continue
		} else if existing, err := c.service.GetById(map[string]interface{}{
			"id": user.GetId(),
		}); err == nil && user.CreatedAt == nil {
			user.CreatedAt = existing[0].CreatedAt
		}
	}

	results, httpErr := c.service.Create(users)
	sanitise(results)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusCreated
	if userErr == nil && len(results) > 0 {
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
	logger.Debugf("user ctrl update\n")

	user := &User{}
	users := []*User{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	userErr, parseErr := reqCtx.Payload.ParseJsonToStruct(user, &users)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if userErr == nil {
		users = append(users, user)
	}

	for _, user := range users {
		if validErr := helper.ValidateStruct(*user); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if user.Id == nil && user.MongoId == nil {
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{"message": "please ensure all records with id for PATCH"},
			)
		}

		conditions := map[string]interface{}{}
		conditions["id"] = user.GetId()

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
		} else if user.CreatedAt == nil {
			user.CreatedAt = existing[0].CreatedAt
		}
	}

	results, httpErr := c.service.Update(users)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}
	sanitise(results)

	respCode = fiber.StatusOK
	if userErr == nil && len(results) > 0 {
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
		results []*User
		err     error
	)

	if cfg.DbConf.Driver == "mongodb" {
		results, err = c.service.Delete(mongoDelIds.Ids)
	} else {
		idsString, _ := helper.ConvertNumberSliceToString(delIds.Ids)
		results, err = c.service.Delete(idsString)
	}
	sanitise(results)

	if err != nil {
		logger.Errorf("failed to delete, err: %+v\n", err.Error())
		respCode = fiber.StatusNotFound
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": err.Error()})
	}

	respCode = fiber.StatusOK
	return fctx.JsonResponse(respCode, map[string]interface{}{"data": results})
}

func (c *Controller) Login(ctx *fiber.Ctx) error {
	mu.Lock() // for avoid sqlite goroute race error
	defer mu.Unlock()

	logger.Debugf("user ctrl login")
	user := &User{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if userErr, _ := reqCtx.Payload.ParseJsonToStruct(user, nil); userErr != nil {
		logger.Errorf("userErr: %+v\n", userErr)
	}
	if user.Password == nil {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": "missing password..."})
	}
	// logger.Debugf("login req: %+v\n", user)

	result, httpErr := c.service.Login(user)
	if httpErr != nil {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": httpErr.Err.Error()})
	}

	if err := SetTokensInCookie(result, ctx); err != nil {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": err.Error()})
	}
	respCode = fiber.StatusOK
	return fctx.JsonResponse(respCode, map[string]interface{}{"data": result})
}

func (c *Controller) Refresh(ctx *fiber.Ctx) error {
	fctx := &helper.FiberCtx{Fctx: ctx}
	// Read cookie
	cookie := ctx.Cookies("refreshToken")

	refreshToken := "Bearer " + cookie
	logger.Debugf("%s\n", refreshToken)

	claims, err := auth.ParseJwt(refreshToken)
	if claims["tokenType"] != "refreshToken" || err != nil {
		respCode = fiber.StatusExpectationFailed
		return fctx.JsonResponse(
			respCode,
			map[string]interface{}{"message": "Invalid Token type... please try to login again"},
		)
	}

	var (
		result     = map[string]interface{}{}
		refreshErr *helper.HttpErr
	)

	if cfg.DbConf.Driver == "mongodb" {
		userId := claims["userId"].(string)
		result, refreshErr = c.service.Refresh(&User{MongoId: &userId})
	} else {
		userId := int64(claims["userId"].(float64))
		// result, refreshErr = c.service.Refresh(&User{Id: &userId})
		result, refreshErr = c.service.Refresh(&User{Id: utils.ToPtr(helper.FlexInt(userId))})
	}
	if refreshErr != nil {
		return fctx.JsonResponse(
			refreshErr.Code,
			map[string]interface{}{"message": refreshErr.Err.Error()},
		)
	}

	if err := SetTokensInCookie(result, ctx); err != nil {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": err.Error()})
	}
	respCode = fiber.StatusOK
	return fctx.JsonResponse(respCode, map[string]interface{}{"data": result})
}

func (c *Controller) LoginPage(ctx *fiber.Ctx) error {
	// data for template
	data := map[string]interface{}{
		"errMessage": nil,
	}
	tpl := template.Must(template.ParseFiles("web/template/login.gohtml", "web/template/base.gohtml"))

	fctx := &helper.FiberCtx{Fctx: ctx}

	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) SendLogin(ctx *fiber.Ctx) error {
	data := fiber.Map{}
	html := `
	<div id="errorMessage" class="mx-auto text-red-600">{{$.message}}</div>
	`
	tpl, _ := template.New("change-password").Parse(html)

	u := new(User)
	fctx := &helper.FiberCtx{Fctx: ctx}

	if err := fctx.Fctx.BodyParser(u); err != nil {
		logger.Errorf("BodyParser err: %+v", err)
		data["message"] = "something went wrong: failed to parse request json"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	result, httpErr := c.service.Login(u)
	if httpErr != nil {
		logger.Errorf("user Login err: %+v", httpErr.Err.Error())
		data["message"] = fmt.Sprintf("login failed: %s", httpErr.Err.Error())
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	if err := SetTokensInCookie(result, ctx); err != nil {
		logger.Errorf("SetTokensInCookie err: %+v", err.Error())
	}

	// login success, redirect to target path/url
	homePage := "/home"
	fctx.Fctx.Set("HX-Redirect", homePage)
	return fctx.Fctx.Redirect(homePage, fiber.StatusOK)
}
