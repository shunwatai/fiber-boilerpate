package todo

import (
	"errors"
	"fmt"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/document"
	"golang-api-starter/internal/modules/todoDocument"
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
	logger.Debugf("todo ctrl\n")
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
	logger.Debugf("todo ctrl\n")
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
	logger.Debugf("todo ctrl create\n")
	c.service.ctx = ctx
	todo := &Todo{}
	todos := []*Todo{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	todoErr, parseErr := reqCtx.Payload.ParseJsonToStruct(todo, &todos)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if todoErr == nil {
		todos = append(todos, todo)
	}
	// logger.Debugf("todoErr: %+v, todosErr: %+v\n", todoErr, todosErr)
	// for _, t := range todos {
	// 	logger.Debugf("todos: %+v\n", t)
	// }

	for _, todo := range todos {
		if validErr := helper.ValidateStruct(*todo); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}

		if todo.Id == nil {
			continue
		} else if existing, err := c.service.GetById(map[string]interface{}{
			"id": todo.GetId(),
		}); err == nil && todo.CreatedAt == nil {
			todo.CreatedAt = existing[0].CreatedAt
		}
		// logger.Debugf("todo? %+v\n", todo)
	}

	// return []*Todo{}
	results, httpErr := c.service.Create(todos)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusCreated
	if todoErr == nil && len(results) > 0 {
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
	logger.Debugf("todo ctrl update\n")

	todo := &Todo{}
	todos := []*Todo{}

	fctx := &helper.FiberCtx{Fctx: ctx}
	reqCtx := &helper.ReqContext{Payload: fctx}
	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": invalidJson.Error()},
		)
	}

	todoErr, parseErr := reqCtx.Payload.ParseJsonToStruct(todo, &todos)
	if parseErr != nil {
		return fctx.JsonResponse(
			fiber.StatusUnprocessableEntity,
			map[string]interface{}{"message": parseErr.Error()},
		)
	}
	if todoErr == nil {
		todos = append(todos, todo)
	}

	for _, todo := range todos {
		if validErr := helper.ValidateStruct(*todo); validErr != nil {
			return fctx.JsonResponse(
				fiber.StatusUnprocessableEntity,
				map[string]interface{}{"message": validErr.Error()},
			)
		}
		if todo.Id == nil && todo.MongoId == nil {
			return fctx.JsonResponse(
				respCode,
				map[string]interface{}{"message": "please ensure all records with id for PATCH"},
			)
		}

		conditions := map[string]interface{}{}
		conditions["id"] = todo.GetId()

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
		} else if todo.CreatedAt == nil {
			todo.CreatedAt = existing[0].CreatedAt
		}
	}

	results, httpErr := c.service.Update(todos)
	if httpErr.Err != nil {
		return fctx.JsonResponse(
			httpErr.Code,
			map[string]interface{}{"message": httpErr.Err.Error()},
		)
	}

	respCode = fiber.StatusOK
	if todoErr == nil && len(results) > 0 {
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
	logger.Debugf("todo ctrl delete\n")
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
		results []*Todo
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

func (c *Controller) ListTodosPage(ctx *fiber.Ctx) error {
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"title":      "Todos",
		"todos":      Todos{},
		"pagination": helper.Pagination{},
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/todos/list.gohtml",
		"web/template/todos/index.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	todos, pagination := c.service.Get(paramsMap)
	data["todos"] = todos
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) GetTodoList(ctx *fiber.Ctx) error {
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"todos":      Todos{},
		"pagination": helper.Pagination{},
	}
	tmplFiles := []string{"web/template/todos/list.gohtml"}
	pagesFunc := helper.TmplCustomFuncs()
	tpl := template.Must(template.New("").Funcs(pagesFunc).ParseFiles(tmplFiles...))
	html := `{{ template "list" . }}`
	tpl, _ = tpl.New("").Parse(html)

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	todos, pagination := c.service.Get(paramsMap)
	data["todos"] = todos
	data["pagination"] = pagination

	fctx := &helper.FiberCtx{Fctx: ctx}
	respCode = fiber.StatusOK

	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	fctx.Fctx.Set("HX-Push-Url", fmt.Sprintf("/todos?%s", string(ctx.Request().URI().QueryString())))
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) TodoFormPage(ctx *fiber.Ctx) error {
	fctx := &helper.FiberCtx{Fctx: ctx}
	// data for template
	data := fiber.Map{
		"errMessage": nil,
		"showNavbar": true,
		"todo":       &Todo{},
		"title":      "Create todo",
	}
	tmplFiles := []string{
		"web/template/parts/popup.gohtml",
		"web/template/todos/form.gohtml",
		"web/template/parts/navbar.gohtml",
		"web/template/base.gohtml",
	}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	paramsMap := helper.GetQueryString(ctx.Request().URI().QueryString())
	u := new(Todo)
	// logger.Debugf("todo_id: %+v", paramsMap["todo_id"])

	if paramsMap["todo_id"] != nil { // update todo
		if cfg.DbConf.Driver == "mongodb" {
			todoId := paramsMap["todo_id"].(string)
			u.MongoId = &todoId
		} else {
			todoId, err := strconv.ParseInt(paramsMap["todo_id"].(string), 10, 64)
			if err != nil {
				return nil
			}

			u.Id = utils.ToPtr(helper.FlexInt(todoId))
		}

		todos, _ := c.service.Get(map[string]interface{}{"id": u.GetId()})
		if len(todos) == 0 {
			logger.Errorf("something went wrong... failed to find todo with id: %+v", u.Id)
			return nil
		}
		data["todo"] = todos[0]
		data["title"] = "Update todo"
	} else { // new todo
		data["todo"] = nil
	}

	respCode = fiber.StatusOK
	fctx.Fctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return tpl.ExecuteTemplate(fctx.Fctx.Response().BodyWriter(), "base.gohtml", data)
}

func (c *Controller) SubmitNew(ctx *fiber.Ctx) error {
	logger.Debugf("todo ctrl form create submit \n")

	respCode = fiber.StatusInternalServerError
	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Response().SetStatusCode(respCode)
	// reqCtx := &helper.ReqContext{Payload: fctx}

	c.service.ctx = ctx
	todo := &Todo{}
	todos := []*Todo{}

	data := fiber.Map{}
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup" . }}`
	tpl, _ = tpl.New("message").Parse(html)

	form, err := fctx.Fctx.MultipartForm()
	if err != nil {
		data["errMessage"] = "failed to parse request formdata"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}
	// logger.Debugf("form value: %+v", form.Value)

	todo.Task = form.Value["task"][0]
	if form.Value["done"][0] == "false" {
		todo.Done = utils.ToPtr(false)
	} else {
		todo.Done = utils.ToPtr(true)
	}

	todos = append(todos, todo)

	for _, todo := range todos {
		if validErr := helper.ValidateStruct(*todo); validErr != nil {
			data["errMessage"] = validErr.Error()
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
	}

	// add new todo record
	todos, httpErr := c.service.Create(todos)
	if httpErr.Err != nil {
		data["errMessage"] = httpErr.Err.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	// add new documents
	document.Srvc.SetCtx(ctx)
	documents, httpErr := document.Srvc.Create(form)
	if httpErr.Err != nil {
		data["errMessage"] = "failed to upload file(s)"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	// add todoDocuments
	todoDocuments := todoDocument.TodoDocuments{}
	for _, document := range documents {
		todoDocuments = append(
			todoDocuments,
			&todoDocument.TodoDocument{TodoId: todos[0].GetId(), DocumentId: document.Id},
		)
	}
	_, httpErr = todoDocument.Srvc.Create(todoDocuments)
	if httpErr.Err != nil {
		data["errMessage"] = "failed to upload file(s)"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	targetPage := "/todos?page=1&items=5"
	fctx.Fctx.Set("HX-Redirect", targetPage)
	respCode = fiber.StatusCreated
	fctx.Fctx.Response().SetStatusCode(respCode)
	return fctx.Fctx.Redirect(targetPage, respCode)
}

func (c *Controller) SubmitUpdate(ctx *fiber.Ctx) error {
	logger.Debugf("todo ctrl form update submit\n")
	respCode = fiber.StatusInternalServerError
	fctx := &helper.FiberCtx{Fctx: ctx}
	fctx.Fctx.Response().SetStatusCode(respCode)
	reqCtx := &helper.ReqContext{Payload: fctx}

	todo := &Todo{}
	todos := []*Todo{}

	data := fiber.Map{}
	tmplFiles := []string{"web/template/parts/popup.gohtml"}
	tpl := template.Must(template.ParseFiles(tmplFiles...))

	html := `{{ template "popup" . }}`
	tpl, _ = tpl.New("message").Parse(html)

	if invalidJson := reqCtx.Payload.ValidateJson(); invalidJson != nil {
		data["errMessage"] = "something went wrong: failed to parse request json"
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	todoErr, parseErr := reqCtx.Payload.ParseJsonToStruct(todo, &todos)
	if parseErr != nil {
		data["errMessage"] = parseErr.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}
	if todoErr == nil {
		todos = append(todos, todo)
	}

	for _, todo := range todos {
		if validErr := helper.ValidateStruct(*todo); validErr != nil {
			data["errMessage"] = validErr.Error()
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
		if todo.Id == nil && todo.MongoId == nil {
			data["errMessage"] = "please ensure all records with id for PATCH"
			return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
		}
	}

	_, httpErr := c.service.Update(todos)
	if httpErr.Err != nil {
		data["errMessage"] = httpErr.Err.Error()
		return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
	}

	fctx.Fctx.Response().SetStatusCode(fiber.StatusOK)
	if len(todos) == 1 {
		targetPage := fmt.Sprintf("/todos?page=1&items=5")
		fctx.Fctx.Set("HX-Redirect", targetPage)
		return nil
	}
	data["successMessage"] = "Update success."
	fctx.Fctx.Set("HX-Trigger", "reloadList")
	return tpl.Execute(fctx.Fctx.Response().BodyWriter(), data)
}

func (c *Controller) SubmitDelete(ctx *fiber.Ctx) error {
	logger.Debugf("todo ctrl form delete submit \n")

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
