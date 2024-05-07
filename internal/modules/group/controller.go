package group

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
	logger.Debugf("group ctrl\n")
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
	logger.Debugf("group ctrl\n")
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
	logger.Debugf("group ctrl create\n")
	c.service.ctx = ctx
	group := &Group{}
	groups := []*Group{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	groupErr, parseErr := reqCtx.Payload.ParseJsonToStruct(group, &groups)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if groupErr == nil {
		groups = append(groups, group)
	}
	// logger.Debugf("groupErr: %+v, groupsErr: %+v\n", groupErr, groupsErr)
	// for _, t := range groups {
	// 	logger.Debugf("groups: %+v\n", t)
	// }

	for _, group := range groups {
		logger.Debugf("group??? %+v", group)
		if validErr := helper.ValidateStruct(*group); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}

		if group.Id == nil {
			continue
		} else if existing, err := c.service.GetById(map[string]interface{}{
			"id": group.GetId(),
		}); err == nil && group.CreatedAt == nil {
			group.CreatedAt = existing[0].CreatedAt
		}
		// logger.Debugf("group? %+v\n", group)
	}

	// return []*Group{}
	results, httpErr := c.service.Create(groups)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusCreated
	if groupErr == nil && len(results) > 0 {
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
	logger.Debugf("group ctrl update\n")

	group := &Group{}
	groups := []*Group{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	groupErr, parseErr := reqCtx.Payload.ParseJsonToStruct(group, &groups)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if groupErr == nil {
		groups = append(groups, group)
	}

	for _, group := range groups {
		if validErr := helper.ValidateStruct(*group); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if group.Id == nil && group.MongoId == nil {
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{"message": "please ensure all records with id for PATCH"},
			)
		}
	}

	results, httpErr := c.service.Update(groups)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusOK
	if groupErr == nil && len(results) > 0 {
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
	logger.Debugf("group ctrl delete\n")
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
		results []*Group
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

func (c *Controller) ListGroupsPage(ctx *fiber.Ctx) error {
	user.Srvc.SetCtx(ctx)
	username := user.Srvc.GetLoggedInUsername()
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"title":      "Groups",
		"groups":     Groups{},
		"pagination": helper.Pagination{},
		"username":   username,
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/groups/list.gohtml",
		"web/template/groups/index.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	groups, pagination := c.service.Get(paramsMap)
	data["groups"] = groups
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) GetGroupList(ctx *fiber.Ctx) error {
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"groups":     Groups{},
		"pagination": helper.Pagination{},
	}
	tmplFiles := []string{"web/template/groups/list.gohtml"}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))
	html := `{{ template "list" . }}`
	tpl, _ = tpl.New("").Parse(html)

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	groups, pagination := c.service.Get(paramsMap)
	data["groups"] = groups
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	fctx.Fctx.Set("HX-Push-Url", fmt.Sprintf("/groups?%s", string(ctx.Request().URI().QueryString())))
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) GroupFormPage(ctx *fiber.Ctx) error {
	user.Srvc.SetCtx(ctx)
	username := user.Srvc.GetLoggedInUsername()
	fctx := &helper.FiberCtx{Fctx: ctx}
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"group":      &Group{},
		"title":      "Create group",
		"username":   username,
		"users":      []user.User{},
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/groups/form.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	u := new(Group)
	// logger.Debugf("group_id: %+v", paramsMap["group_id"])

	if paramsMap["group_id"] != nil { // update group
		if cfg.DbConf.Driver == "mongodb" {
			groupId := paramsMap["group_id"].(string)
			u.MongoId = &groupId
		} else {
			groupId, err := strconv.ParseInt(paramsMap["group_id"].(string), 10, 64)
			if err != nil {
				return nil
			}

			u.Id = utils.ToPtr(helper.FlexInt(groupId))
		}

		groups, _ := c.service.Get(map[string]interface{}{"id": u.GetId()})
		if len(groups) == 0 {
			logger.Errorf("something went wrong... failed to find group with id: %+v", u.Id)
			return nil
		}
		users, _ := user.Srvc.Get(map[string]interface{}{"disabled": false})
		userIdMap := user.Srvc.GetIdMap(groups[0].Users)
		usersWithoutExisting := []*user.User{}
		for _, u := range users {
			_, ok := userIdMap[u.GetId()]
			if ok {
				continue
			}
			usersWithoutExisting = append(usersWithoutExisting, u)
		}
		data["group"] = groups[0]
		data["users"] = usersWithoutExisting
		data["title"] = "Update group"
	} else { // new group
		data["group"] = nil
	}

	respCode = fiber.StatusOK
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) SubmitNew(ctx *fiber.Ctx) error {
	logger.Debugf("group ctrl form create submit \n")

	respCode = fiber.StatusInternalServerError
	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Response().SetStatusCode(respCode)
	reqCtx := &helper.ReqContext{Payload: fctx}

	c.service.ctx = ctx
	group := &Group{}
	groups := []*Group{}

	data := fiber.Map{}
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup" . }}`
	tpl, _ = tpl.New("message").Parse(html)

	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		data["errMessage"] = invalidJson.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	groupErr, parseErr := reqCtx.Payload.ParseJsonToStruct(group, &groups)
	if parseErr != nil {
		data["errMessage"] = parseErr.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}
	if groupErr == nil {
		groups = append(groups, group)
	}

	for _, group := range groups {
		if validErr := helper.ValidateStruct(*group); validErr != nil {
			data["errMessage"] = validErr.Error()
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
	}

	_, httpErr := c.service.Create(groups)
	if httpErr.Err != nil {
		data["errMessage"] = httpErr.Err.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	targetPage := "/groups?page=1&items=5"
	fctx.Fctx.Set("HX-Redirect", targetPage)
	respCode = fiber.StatusCreated
	fctx.Fctx.Response().SetStatusCode(respCode)
	return fctx.Fctx.Redirect(targetPage, respCode)
}

func (c *Controller) SubmitUpdate(ctx *fiber.Ctx) error {
	logger.Debugf("group ctrl form update submit\n")
	respCode = fiber.StatusInternalServerError
	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Response().SetStatusCode(respCode)
	reqCtx := &helper.ReqContext{Payload: fctx}

	group := &Group{}
	groups := []*Group{}

	data := fiber.Map{}
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup" . }}`
	tpl, _ = tpl.New("message").Parse(html)

	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		data["errMessage"] = "something went wrong: failed to parse request json"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	groupErr, parseErr := reqCtx.Payload.ParseJsonToStruct(group, &groups)
	if parseErr != nil {
		data["errMessage"] = parseErr.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}
	if groupErr == nil {
		groups = append(groups, group)
	}

	for _, group := range groups {
		if validErr := helper.ValidateStruct(*group); validErr != nil {
			data["errMessage"] = validErr.Error()
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
		if group.Id == nil && group.MongoId == nil {
			data["errMessage"] = "please ensure all records with id for PATCH"
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
	}

	_, httpErr := c.service.Update(groups)
	if httpErr.Err != nil {
		data["errMessage"] = httpErr.Err.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	fctx.Fctx.Response().SetStatusCode(fiber.StatusOK)
	if len(groups) == 1 {
		targetPage := fmt.Sprintf("/groups?page=1&items=5")
		fctx.Fctx.Set("HX-Redirect", targetPage)
		return nil
	}
	data["successMessage"] = "Update success."
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) SubmitDelete(ctx *fiber.Ctx) error {
	logger.Debugf("group ctrl form delete submit \n")

	respCode = fiber.StatusInternalServerError
	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Response().SetStatusCode(respCode)
	reqCtx := &helper.ReqContext{Payload: fctx}

	c.service.ctx = ctx

	data := fiber.Map{}
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup" . }}`
	tpl, _ = tpl.New("message").Parse(html)

	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		data["errMessage"] = invalidJson.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	delIds := struct {
		Ids []helper.FlexInt `json:"ids" validate:"required,unique"`
	}{}

	mongoDelIds := struct {
		Ids []string `json:"ids" validate:"required,unique"`
	}{}

	intIdsErr, strIdsErr := reqCtx.Payload.ParseJsonToStruct(&delIds, &mongoDelIds)
	if intIdsErr != nil && strIdsErr != nil {
		logger.Errorf("failed to parse req json, %+v\n", errors.Join(intIdsErr, strIdsErr).Error())
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": errors.Join(intIdsErr, strIdsErr).Error()})
	}
	if len(delIds.Ids) == 0 && len(mongoDelIds.Ids) == 0 {
		return fctx.JsonResponse(respCode, map[string]interface{}{"message": "please check the req json like the follow: {\"ids\":[]}"})
	}
	logger.Debugf("deletedIds: %+v, mongoIds: %+v\n", delIds, mongoDelIds)

	var err error

	if cfg.DbConf.Driver == "mongodb" {
		_, err = c.service.Delete(mongoDelIds.Ids)
	} else {
		idsString, _ := helper.ConvertNumberSliceToString(delIds.Ids)
		_, err = c.service.Delete(idsString)
	}
	if err != nil {
		data["errMessage"] = err.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	fctx.Fctx.Response().SetStatusCode(fiber.StatusNoContent)
	fctx.Fctx.Set("HX-Refresh", "true")
	return nil
}
