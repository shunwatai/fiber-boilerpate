package todoDocument

import (
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/middleware/jwtcheck"

	"github.com/gofiber/fiber/v2"
)

var (
	cfg               = config.Cfg
	tableName         = "todo_documents"
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

	r := router.Group("/api/todo-documents", jwtcheck.CheckJwt())
	r.Get("/", GetAll)
	r.Post("/", Create)
	r.Patch("/", Update)
	r.Delete("/", Delete)

	rById := r.Group("/:id")
	rById.Get("/", GetById)
}

// TodoDocumentGetAll godoc
//
//	@Summary		List TodoDocuments
//	@Description	get TodoDocuments
//	@Tags			todoDocuments
//	@Accept			json
//	@Produce		json
//	@Param			id		query	number	false	"id"							example(2)
//	@Param			userId	query	number	false	"search by userId"				example(2)
//	@Param			task	query	string	false	"search by task"				example(go practice)
//	@Param			done	query	boolean	false	"search by done"				example(1)
//	@Param			page	query	string	false	"page number for pagination"	example(1)
//	@Param			items	query	string	false	"items per page for pagination"	example(10)
//	@Security		ApiKeyAuth
//	@Router			/todo-documents [get]
func GetAll(c *fiber.Ctx) error {
	return ctrl.Get(c)
}

// GetTodoDocumentById godoc
//
//	@Summary		Get TodoDocument by ID
//	@Description	get TodoDocument by ID
//	@Tags			todoDocuments
//	@Accept			json
//	@Produce		json
//	@Param			todoDocumentId	path	int	true	"todoDocument ID"	example(12)
//	@Security		ApiKeyAuth
//	@Router			/todo-documents/{todoDocumentId} [get]
func GetById(c *fiber.Ctx) error {
	return ctrl.GetById(c)
}

// TodoDocumentCreate godoc
//
//	@Summary		Create new todoDocument(s)
//	@Description	Create todoDocument(s)
//	@Tags			todoDocuments
//	@Accept			json
//	@Produce		json
//	@Param			TodoDocument	body	string	true	"single TodoDocument request json"	SchemaExample({ "task": "take shower", "done": false })
//	@Param			TodoDocuments	body	string	true	"batch TodoDocument request json"	SchemaExample([{ "task": "take shower", "done": false }, { "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/todo-documents [post]
func Create(c *fiber.Ctx) error {
	return ctrl.Create(c)
}

// TodoDocumentUpdate godoc
//
//	@Summary		Update existing todoDocument(s)
//	@Description	Update todoDocument(s)
//	@Tags			todoDocuments
//	@Accept			json
//	@Produce		json
//	@Param			TodoDocument	body	string	true	"single TodoDocument request json"	SchemaExample({ "id":2, "task": "take shower", "done": false })
//	@Param			TodoDocuments	body	string	true	"batch TodoDocument request json"	SchemaExample([{ "id":2, "task": "take shower", "done": false, createdAt: "2021-01-11" }, { "id":3, "task": "go practice", "done": false }])
//	@Security		ApiKeyAuth
//	@Router			/todo-documents [patch]
func Update(c *fiber.Ctx) error {
	return ctrl.Update(c)
}

// RemoveTodoDocument godoc
//
//	@Summary		Delete todoDocument(s)
//	@Description	delete todoDocument(s)
//	@Tags			todoDocuments
//	@Accept			json
//	@Produce		json
//	@Param			todoDocumentIds	body	string	true	"array of todoDocument IDs"	SchemaExample([1,2,3])
//	@Security		ApiKeyAuth
//	@Router			/todo-documents [delete]
func Delete(c *fiber.Ctx) error {
	return ctrl.Delete(c)
}
