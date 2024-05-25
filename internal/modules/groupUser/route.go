package groupUser

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg               = config.Cfg
	tableName         = "group_users"
	viewName  *string = nil
	Repo              = &Repository{}
	Srvc              = &Service{}
	ctrl              = &Controller{}
)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares, groupRepo IGroupRepository, userRepo IUserRepository) {
	db := database.GetDatabase(tableName, viewName)
	Repo = NewRepository(db)
	Repo.GroupRepo = groupRepo
	Repo.UserRepo = userRepo
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	r := router.Group("/api/group-users", custMiddleware.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// GroupUserGetAll godoc
//
//	@Summary		List GroupUsers
//	@Description	get GroupUsers
//	@Tags			groupUsers
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/group-users [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetGroupUserById godoc
//
//	@Summary		Get GroupUser by ID
//	@Description	get GroupUser by ID
//	@Tags			groupUsers
//	@Accept			json
//	@Produce		json
//	@Param			groupUserId	path	int	true	"groupUser ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/group-users/{groupUserId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// GroupUserCreate godoc
//
//	@Summary		Create new groupUser(s)
//	@Description	Create groupUser(s)
//	@Tags			groupUsers
//	@Accept			json
//	@Produce		json
//	@Param			GroupUser	body	string	true	"single GroupUser request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			GroupUsers	body	string	true	"batch GroupUser request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/group-users [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// GroupUserUpdate godoc
//
//	@Summary		Update existing groupUser(s)
//	@Description	Update groupUser(s)
//	@Tags			groupUsers
//	@Accept			json
//	@Produce		json
//	@Param			GroupUser	body	string	true	"single GroupUser request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			GroupUsers	body	string	true	"batch GroupUser request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/group-users [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveGroupUser godoc
//
//	@Summary		Delete groupUser(s)
//	@Description	delete groupUser(s)
//	@Tags			groupUsers
//	@Accept			json
//	@Produce		json
//	@Param			groupUserIds	body	string	true	"array of groupUser IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/group-users [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
