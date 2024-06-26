package log

import (
	"errors"
	"fmt"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/user"
	"html/template"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	service *Service
}

func NewController(s *Service) *Controller {
	return &Controller{s}
}

var respCode = fiber.StatusInternalServerError

func (c *Controller) Get(ctx *fiber.Ctx) error {
	logger.Debugf("log ctrl\n")
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
	logger.Debugf("log ctrl\n")
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
	logger.Debugf("log ctrl create\n")
	c.service.ctx = ctx
	log := &Log{}
	logs := []*Log{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	logErr, parseErr := reqCtx.Payload.ParseJsonToStruct(log, &logs)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if logErr == nil {
		logs = append(logs, log)
	}
	// logger.Debugf("logErr: %+v, logsErr: %+v\n", logErr, logsErr)
	// for _, t := range logs {
	// 	logger.Debugf("logs: %+v\n", t)
	// }

	for _, log := range logs {
		if validErr := helper.ValidateStruct(*log); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}

		if log.Id == nil {
			continue
		} else if existing, err := c.service.GetById(map[string]interface{}{
			"id": log.GetId(),
		}); err == nil && log.CreatedAt == nil {
			log.CreatedAt = existing[0].CreatedAt
		}
		// logger.Debugf("log? %+v\n", log)
	}

	// return []*Log{}
	results, httpErr := c.service.Create(logs)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusCreated
	if logErr == nil && len(results) > 0 {
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
	logger.Debugf("log ctrl update\n")

	log := &Log{}
	logs := []*Log{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	logErr, parseErr := reqCtx.Payload.ParseJsonToStruct(log, &logs)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if logErr == nil {
		logs = append(logs, log)
	}

	for _, log := range logs {
		if validErr := helper.ValidateStruct(*log); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if log.Id == nil && log.MongoId == nil {
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{"message": "please ensure all records with id for PATCH"},
			)
		}
	}

	results, httpErr := c.service.Update(logs)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusOK
	if logErr == nil && len(results) > 0 {
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
	logger.Debugf("log ctrl delete\n")
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
		results []*Log
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

/* view controllers start here */
func (c *Controller) ListLogsPage(ctx *fiber.Ctx) error {
	user.Srvc.SetCtx(ctx)
	username := user.Srvc.GetLoggedInUsername()
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"title":      "Logs",
		"logs":       Logs{},
		"pagination": helper.Pagination{},
		"logname":    username,
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/logs/list.gohtml",
		"web/template/logs/index.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	logs, pagination := c.service.Get(paramsMap)
	logger.Debugf("logs?? %+v", logs)
	data["logs"] = logs
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) GetLogList(ctx *fiber.Ctx) error {
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"logs":       Logs{},
		"pagination": helper.Pagination{},
	}
	tmplFiles := []string{"web/template/logs/list.gohtml"}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))
	html := `{{ template "list" . }}`
	tpl, _ = tpl.New("").Parse(html)

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	logs, pagination := c.service.Get(paramsMap)
	data["logs"] = logs
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	fctx.Fctx.Set("HX-Push-Url", fmt.Sprintf("/logs?%s", string(ctx.Request().URI().QueryString())))
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) LogDetailPage(ctx *fiber.Ctx) error {
	user.Srvc.SetCtx(ctx)
	username := user.Srvc.GetLoggedInUsername()
	fctx := &helper.FiberCtx{Fctx: ctx}
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"log":        &Log{},
		"title":      "Log detail",
		"username":   username,
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/logs/form.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	lg := new(Log)
	// logger.Debugf("log_id: %+v", paramsMap["log_id"])
	logger.Debugf("log_id: %+v", paramsMap["log_id"])

	if cfg.DbConf.Driver == "mongodb" {
		logId := paramsMap["log_id"].(string)
		lg.MongoId = &logId
	} else {
		logId, err := strconv.ParseInt(paramsMap["log_id"].(string), 10, 64)
		if err != nil {
			return nil
		}

		lg.Id = utils.ToPtr(helper.FlexInt(logId))

		// get group by ID
		logs, _ := c.service.Get(map[string]interface{}{"id": lg.GetId()})
		if len(logs) == 0 {
			logger.Errorf("something went wrong... failed to find log with id: %+v", lg.Id)
			return nil
		}

		data["log"] = logs[0]
	}

	respCode = fiber.StatusOK
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}
