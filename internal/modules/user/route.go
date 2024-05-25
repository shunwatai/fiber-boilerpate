package user

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"
	"golang-api-starter/internal/modules/groupUser"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg       = config.Cfg
	tableName = "users"
	viewName  = "users_view"
	Repo      = &Repository{}
	Srvc      = &Service{}
	ctrl      = &Controller{}
)

func GetRoutes(router fiber.Router, custMiddleware interfaces.ICustomMiddlewares, groupRepo groupUser.IGroupRepository ) {
	db := database.GetDatabase(tableName, &viewName)
	Repo = NewRepository(db)
	Repo.GroupRepo = groupRepo
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	// normal auth from database's users table
	authRoute := router.Group("/api/auth")
	authRoute.Post("/login", Login)
	authRoute.Get("/logout", ctrl.Logout)
	authRoute.Post("/refresh", Refresh)

	// web view routes
	publicViewRoute := router.Group("")
	publicViewRoute.Get("/login", ctrl.LoginPage)
	publicViewRoute.Post("/login", ctrl.SubmitLogin)

	protectedViewRoute := router.Group("/users", custMiddleware.CheckJwt(), custMiddleware.CheckAccess("users"))
	// protectedViewRoute := router.Group("/users", custMiddleware.CheckJwt())
	protectedViewRoute.Route("", func(userPage fiber.Router) {
		userPage.Get("/", ctrl.ListUsersPage)
		userPage.Get("/list", ctrl.GetUserList)
		userPage.Delete("/", ctrl.SubmitDelete)
		userPage.Patch("/", ctrl.SubmitUpdate)
		userPage.Post("/", ctrl.SubmitNew)
		userPage.Route("/form", func(userForm fiber.Router) {
			userForm.Get("/", ctrl.UserFormPage)
		})
	})

	// users routes
	r := router.Group("/api/users", custMiddleware.CheckJwt(), custMiddleware.CheckAccess("users"))
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// UserGetAll godoc
//
//	@Summary		List Users
//	@Description	get Users
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id			query	number	false	"id"							example(2)
//	@Param			name		query	string	false	"search by name"				example(tom)
//	@Param			firstName	query	string	false	"search by firstName"			example(will)
//	@Param			lastName	query	string	false	"search by lastName"			example(smith)
//	@Param			disabled	query	boolean	false	"search by disabled"			example(0)
//	@Param			page		query	string	false	"page number for pagination"	example(1)
//	@Param			items		query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/users [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetUserById godoc
//
//	@Summary		Get User by ID
//	@Description	get User by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path	int	true	"user ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/users/{userId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// UserCreate godoc
//
//	@Summary		Create new user(s)
//	@Description	Create user(s)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			User	body	string	true	"single User request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			Users	body	string	true	"batch User request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/users [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// UserUpdate godoc
//
//	@Summary		Update existing user(s)
//	@Description	Update user(s)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			User	body	string	true	"single User request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			Users	body	string	true	"batch User request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/users [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveUser godoc
//
//	@Summary		Delete user(s)
//	@Description	delete user(s)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userIds	body	string	true	"array of user IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/users [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}

// UserLogin godoc
//
//	@Summary		Login user
//	@Description	login user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			User	body	string	true	"Login request json"	SchemaExample({ "name": "admin", "password": "admin" })
//	@Security		ApiKeyAuth
//	@Router			/auth/login [post]
func Login(c *fiber.Ctx) error {
	return ctrl.Login(c)
}

// UserRefresh godoc
//
//	@Summary		Refrese user
//	@Description	refresh user
//	@Tags			users
//	@Router			/auth/refresh [post]
func Refresh(c *fiber.Ctx) error {
	return ctrl.Refresh(c)
}
