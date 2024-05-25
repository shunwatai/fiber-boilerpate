package resource

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg       = config.Cfg
	tableName = "resources"
	viewName  *string = nil
	Repo      = &Repository{}
	Srvc      = &Service{}
	ctrl      = &Controller{}
)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares) {
	db := database.GetDatabase(tableName, viewName)
	Repo = NewRepository(db)
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	r := router.Group("/api/resources", custMiddleware.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// ResourceGetAll godoc
//
//	@Summary		List Resources
//	@Description	get Resources
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/resources [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetResourceById godoc
//
//	@Summary		Get Resource by ID
//	@Description	get Resource by ID
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			resourceId	path	int	true	"resource ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/resources/{resourceId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// ResourceCreate godoc
//
//	@Summary		Create new resource(s)
//	@Description	Create resource(s)
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			Resource	body	string	true	"single Resource request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			Resources	body	string	true	"batch Resource request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/resources [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// ResourceUpdate godoc
//
//	@Summary		Update existing resource(s)
//	@Description	Update resource(s)
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			Resource	body	string	true	"single Resource request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			Resources	body	string	true	"batch Resource request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/resources [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveResource godoc
//
//	@Summary		Delete resource(s)
//	@Description	delete resource(s)
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			resourceIds	body	string	true	"array of resource IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/resources [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
