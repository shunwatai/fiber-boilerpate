package todo

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/interfaces"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg               = config.Cfg
	tableName         = "todos"
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

	protectedViewRoute := router.Group("/todos", custMiddleware.CheckJwt(), custMiddleware.CheckAccess("todos"))
	protectedViewRoute.Route("", func(todoPage fiber.Router) {
		todoPage.Get("/", ctrl.ListTodosPage)
		todoPage.Get("/list", ctrl.GetTodoList)
		todoPage.Delete("/", ctrl.SubmitDelete)
		todoPage.Patch("/", ctrl.SubmitUpdate)
		todoPage.Patch("/toggle-done", ctrl.ToggleDone)
		todoPage.Post("/", ctrl.SubmitNew)
		todoPage.Route("/form", func(todoForm fiber.Router) {
			todoForm.Get("/", ctrl.TodoFormPage)
		})
	})

	r := router.Group("/api/todos", custMiddleware.CheckJwt(), custMiddleware.CheckAccess("todos"))
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// TodoGetAll godoc
//
//	@Summary		List Todos
//	@Description	get Todos
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/todos [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetTodoById godoc
//
//	@Summary		Get Todo by ID
//	@Description	get Todo by ID
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Param			todoId	path	int	true	"todo ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/todos/{todoId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// TodoCreate godoc
//
//	@Summary		Create new todo(s)
//	@Description	Create todo(s)
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Param			Todo	body	string	true	"single Todo request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			Todos	body	string	true	"batch Todo request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/todos [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// TodoUpdate godoc
//
//	@Summary		Update existing todo(s)
//	@Description	Update todo(s)
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Param			Todo	body	string	true	"single Todo request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			Todos	body	string	true	"batch Todo request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/todos [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveTodo godoc
//
//	@Summary		Delete todo(s)
//	@Description	delete todo(s)
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Param			todoIds	body	string	true	"array of todo IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/todos [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
