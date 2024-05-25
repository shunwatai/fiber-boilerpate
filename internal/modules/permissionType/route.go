package permissionType

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg       = config.Cfg
	tableName = "permission_types"
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

	r := router.Group("/api/permission-types", custMiddleware.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// PermissionTypeGetAll godoc
//
//	@Summary		List PermissionTypes
//	@Description	get PermissionTypes
//	@Tags			permissionTypes
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/permission-types [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetPermissionTypeById godoc
//
//	@Summary		Get PermissionType by ID
//	@Description	get PermissionType by ID
//	@Tags			permissionTypes
//	@Accept			json
//	@Produce		json
//	@Param			permissionTypeId	path	int	true	"permissionType ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/permission-types/{permissionTypeId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// PermissionTypeCreate godoc
//
//	@Summary		Create new permissionType(s)
//	@Description	Create permissionType(s)
//	@Tags			permissionTypes
//	@Accept			json
//	@Produce		json
//	@Param			PermissionType	body	string	true	"single PermissionType request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			PermissionTypes	body	string	true	"batch PermissionType request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/permission-types [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// PermissionTypeUpdate godoc
//
//	@Summary		Update existing permissionType(s)
//	@Description	Update permissionType(s)
//	@Tags			permissionTypes
//	@Accept			json
//	@Produce		json
//	@Param			PermissionType	body	string	true	"single PermissionType request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			PermissionTypes	body	string	true	"batch PermissionType request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/permission-types [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemovePermissionType godoc
//
//	@Summary		Delete permissionType(s)
//	@Description	delete permissionType(s)
//	@Tags			permissionTypes
//	@Accept			json
//	@Produce		json
//	@Param			permissionTypeIds	body	string	true	"array of permissionType IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/permission-types [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
