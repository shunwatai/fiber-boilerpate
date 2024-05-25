package oauth

import (
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/groupUser"
	"golang-api-starter/internal/modules/user"
	"html/template"

	"github.com/gofiber/fiber/v2"
	fiberUtils "github.com/gofiber/fiber/v2/utils"
	"github.com/shareed2k/goth_fiber"
)

type Controller struct {
	service *Service
}

func NewController(s *Service) *Controller {
	return &Controller{s}
}

var respCode = fiber.StatusInternalServerError

func (c *Controller) OAuthGetAuth(ctx *fiber.Ctx) error {
	oauthUser, err := goth_fiber.CompleteUserAuth(ctx)
	if err != nil {
		logger.Errorf("goth_fiber.CompleteUserAuth err: %+v", err)
		return err
	}

	provider, _ := goth_fiber.GetProviderName(ctx)
	logger.Debugf("provider: %+v", provider)

	logger.Debugf("authed user: %+v", oauthUser)
	username := ""
	if len(oauthUser.Email) > 0 {
		username = oauthUser.Email
	} else if len(oauthUser.NickName) > 0 {
		username = oauthUser.NickName
	} else if len(oauthUser.Name) > 0 {
		username = oauthUser.Name
	} else {
		return logger.Errorf("failed to get the username from oauthUser")
	}
	logger.Debugf("username: %+v", username)

	users, _ := user.Srvc.Get(map[string]interface{}{"name": username, "exactMatch": map[string]bool{"name": true}})
	if len(users) == 0 { // new user, add to db
		users = append(users, &groupUser.User{
			Name:     username,
			Password: utils.ToPtr(fiberUtils.UUIDv4()), // useless random dummy password for oauth user
			IsOauth:  true,
			Provider: utils.ToPtr(oauthUser.Provider),
		})
		user.Srvc.Create(users)
	}

	// return jwt token
	// logger.Debugf("users[0]: %+v", users[0])
	result, httpErr := user.Srvc.Login(users[0])

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK
	if httpErr != nil {
		logger.Errorf("user.Srvc.Login err: %+v", err)
		return fctx.JsonResponse(respCode, fiber.Map{"message": httpErr.Err.Error()})
	}

	return fctx.JsonResponse(
		respCode,
		// fiber.Map{"data": oauthUser},
		fiber.Map{"data": result},
	)
}

func (c *Controller) OAuthLogin(ctx *fiber.Ctx) error {
	if err := goth_fiber.BeginAuthHandler(ctx); err != nil {
		logger.Errorf("goth_fiber.BeginAuthHandler err: %+v", err)
	}
	return nil
}

func (c *Controller) OAuthLogout(ctx *fiber.Ctx) error {
	respCode = fiber.StatusOK
	fctx := &helper.FiberCtx{Fctx: ctx}

	provider, _ := goth_fiber.GetProviderName(ctx)
	logger.Debugf("provider: %+v", provider)
	if err := goth_fiber.Logout(ctx); err != nil {
		logger.Errorf("failed to logout...")
	}

	return fctx.JsonResponse(
		respCode,
		fiber.Map{"data": "logged out"},
	)
}

func (c *Controller) OAuthProviderPage(ctx *fiber.Ctx) error {
	tpl := template.Must(template.ParseGlob("web/template/oauth/sign-in.gohtml"))

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), struct{}{})
}
