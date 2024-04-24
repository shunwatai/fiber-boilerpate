package passwordReset

import (
	"github.com/gofiber/fiber/v2"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/middleware/jwtcheck"
)

var (
	cfg               = config.Cfg
	tableName         = "password_resets"
	viewName  *string = nil
	Repo              = &Repository{}
	Srvc              = &Service{}
	ctrl              = &Controller{}
)

func GetRoutes(router fiber.Router) {
	db := database.GetDatabase(tableName, viewName)
	Repo = NewRepository(db)
	Srvc = NewService(Repo)
	ctrl = NewController(Srvc)

	viewRoute := router.Group("/password-resets")
	viewRoute.Get("/", ctrl.PasswordResetPage)
	viewRoute.Get("/forgot", ctrl.SendResetEmailPage)
	viewRoute.Post("/send", ctrl.SendResetEmail)
	viewRoute.Patch("/", ctrl.ChangePassword, jwtcheck.CheckJwt())

	r := router.Group("/api/password-resets", jwtcheck.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// PasswordResetGetAll godoc
//
//	@Summary		List PasswordResets
//	@Description	get PasswordResets
//	@Tags			passwordResets
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/password-resets [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetPasswordResetById godoc
//
//	@Summary		Get PasswordReset by ID
//	@Description	get PasswordReset by ID
//	@Tags			passwordResets
//	@Accept			json
//	@Produce		json
//	@Param			passwordResetId	path	int	true	"passwordReset ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/password-resets/{passwordResetId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// PasswordResetCreate godoc
//
//	@Summary		Create new passwordReset(s)
//	@Description	Create passwordReset(s)
//	@Tags			passwordResets
//	@Accept			json
//	@Produce		json
//	@Param			PasswordReset	body	string	true	"single PasswordReset request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			PasswordResets	body	string	true	"batch PasswordReset request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/password-resets [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// PasswordResetUpdate godoc
//
//	@Summary		Update existing passwordReset(s)
//	@Description	Update passwordReset(s)
//	@Tags			passwordResets
//	@Accept			json
//	@Produce		json
//	@Param			PasswordReset	body	string	true	"single PasswordReset request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			PasswordResets	body	string	true	"batch PasswordReset request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/password-resets [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemovePasswordReset godoc
//
//	@Summary		Delete passwordReset(s)
//	@Description	delete passwordReset(s)
//	@Tags			passwordResets
//	@Accept			json
//	@Produce		json
//	@Param			passwordResetIds	body	string	true	"array of passwordReset IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/password-resets [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
