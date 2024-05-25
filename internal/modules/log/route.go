package log

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg               = config.Cfg
	tableName         = "logs"
	viewName  *string = nil
	Repo              = &Repository{}
	Srvc              = &Service{}
	ctrl              = &Controller{}
)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares) {
	db := database.GetDatabase(tableName, viewName)
	Repo = NewRepository(db)
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	r := router.Group("/api/logs", custMiddleware.CheckJwt())
	r.Get("/", GetAll)
	// r.Post("/", Create)
	// r.Patch("/", Update)
	// r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// LogGetAll godoc
//
//	@Summary		List Logs
//	@Description	get Logs
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/logs [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetLogById godoc
//
//	@Summary		Get Log by ID
//	@Description	get Log by ID
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			logId	path	int	true	"log ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/logs/{logId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// LogCreate godoc
//
//	@Summary		Create new log(s)
//	@Description	Create log(s)
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			Log		body	string	true	"single Log request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			Logs	body	string	true	"batch Log request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/logs [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// LogUpdate godoc
//
//	@Summary		Update existing log(s)
//	@Description	Update log(s)
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			Log		body	string	true	"single Log request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			Logs	body	string	true	"batch Log request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/logs [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveLog godoc
//
//	@Summary		Delete log(s)
//	@Description	delete log(s)
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			logIds	body	string	true	"array of log IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/logs [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
