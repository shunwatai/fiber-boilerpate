package group

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/middleware/jwtcheck"
	"github.com/gofiber/fiber/v2"
)

var (
	cfg       = config.Cfg
	tableName = "groups"
	viewName  *string = nil
	Repo      = &Repository{}
	Srvc      = &Service{}
	ctrl      = &Controller{}
)

func GetRoutes(router fiber.Router) {
	db := database.GetDatabase(tableName, viewName)
	Repo = NewRepository(db)
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	r := router.Group("/api/groups", jwtcheck.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// GroupGetAll godoc
//
//	@Summary		List Groups
//	@Description	get Groups
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/groups [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetGroupById godoc
//
//	@Summary		Get Group by ID
//	@Description	get Group by ID
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			groupId	path	int	true	"group ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/groups/{groupId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// GroupCreate godoc
//
//	@Summary		Create new group(s)
//	@Description	Create group(s)
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			Group	body	string	true	"single Group request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			Groups	body	string	true	"batch Group request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/groups [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// GroupUpdate godoc
//
//	@Summary		Update existing group(s)
//	@Description	Update group(s)
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			Group	body	string	true	"single Group request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			Groups	body	string	true	"batch Group request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/groups [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveGroup godoc
//
//	@Summary		Delete group(s)
//	@Description	delete group(s)
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			groupIds	body	string	true	"array of group IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/groups [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
