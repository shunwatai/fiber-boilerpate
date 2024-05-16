package groupResourceAcl

import (
	"github.com/gofiber/fiber/v2"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/middleware/jwtcheck"
)

var (
	cfg       = config.Cfg
	tableName = "group_resource_acls"
	viewName  = "group_resource_acls_view"
	Repo      = &Repository{}
	Srvc      = &Service{}
	ctrl      = &Controller{}
)

func GetRoutes(router fiber.Router) {
	db := database.GetDatabase(tableName, &viewName)
	Repo = NewRepository(db)
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	r := router.Group("/api/group-resource-acls", jwtcheck.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// GroupResourceAclGetAll godoc
//
//	@Summary		List GroupResourceAcls
//	@Description	get GroupResourceAcls
//	@Tags			groupResourceAcls
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/group-resource-acls [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetGroupResourceAclById godoc
//
//	@Summary		Get GroupResourceAcl by ID
//	@Description	get GroupResourceAcl by ID
//	@Tags			groupResourceAcls
//	@Accept			json
//	@Produce		json
//	@Param			groupResourceAclId	path	int	true	"groupResourceAcl ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/group-resource-acls/{groupResourceAclId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// GroupResourceAclCreate godoc
//
//	@Summary		Create new groupResourceAcl(s)
//	@Description	Create groupResourceAcl(s)
//	@Tags			groupResourceAcls
//	@Accept			json
//	@Produce		json
//	@Param			GroupResourceAcl	body	string	true	"single GroupResourceAcl request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			GroupResourceAcls	body	string	true	"batch GroupResourceAcl request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/group-resource-acls [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// GroupResourceAclUpdate godoc
//
//	@Summary		Update existing groupResourceAcl(s)
//	@Description	Update groupResourceAcl(s)
//	@Tags			groupResourceAcls
//	@Accept			json
//	@Produce		json
//	@Param			GroupResourceAcl	body	string	true	"single GroupResourceAcl request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			GroupResourceAcls	body	string	true	"batch GroupResourceAcl request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/group-resource-acls [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveGroupResourceAcl godoc
//
//	@Summary		Delete groupResourceAcl(s)
//	@Description	delete groupResourceAcl(s)
//	@Tags			groupResourceAcls
//	@Accept			json
//	@Produce		json
//	@Param			groupResourceAclIds	body	string	true	"array of groupResourceAcl IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/group-resource-acls [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
