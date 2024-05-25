package todo

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/document"
	"golang-api-starter/internal/modules/groupUser"
	"golang-api-starter/internal/modules/todoDocument"
	"golang-api-starter/internal/modules/user"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db database.IDatabase
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db}
}

func cascadeFields(todos Todos) {
	if len(todos) == 0 {
		return
	}
	// cascade user
	var (
		userIds []string
		userId  string
	)
	// get all userIds
	for _, todo := range todos {
		if todo.UserId == nil {
			continue
		}

		userId = todo.GetUserId()
		userIds = append(userIds, userId)
	}

	// if no userIds, do nothing and return
	if len(userIds) > 0 {
		users := []*groupUser.User{}

		// get users by userIds
		condition := database.GetIdsMapCondition(nil, userIds)
		users, _ = user.Srvc.Get(condition)
		// get the map[userId]user
		userMap := user.Repo.GetIdMap(users)

		for _, todo := range todos {
			if todo.UserId == nil {
				continue
			}
			user := &groupUser.User{}
			// take out the user by userId in map and assign
			user = userMap[todo.GetUserId()]
			todo.User = user
		}
	}

	// cascade todo-documents
	var (
		todoIds []string
		todoId  string
	)
	// get all todoId
	for _, todo := range todos {
		todoId = todo.GetId()
		todoIds = append(todoIds, todoId)
	}

	todoDocuments := []*todoDocument.TodoDocument{}
	// get users by userIds
	condition := database.GetIdsMapCondition(utils.ToPtr("todo_id"), todoIds)
	todoDocuments, _ = todoDocument.Srvc.Get(condition)

	// get the map[userId]user
	todoDocumentsMap := todoDocument.Srvc.GetTodoIdMap(todoDocuments)

	for _, todo := range todos {
		tds := []*todoDocument.TodoDocument{}
		// take out the user by userId in map and assign
		tds, haveDocuments := todoDocumentsMap[todo.GetId()]

		// if no documents assign empty slice for response json "documents": [] instead of "documents": null
		if !haveDocuments {
			todo.Documents = []*document.Document{}
		} else {
			todo.TodoDocuments = tds
			for _, td := range tds {
				todo.Documents = append(todo.Documents, td.Document)
			}
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*Todo, *helper.Pagination) {
	logger.Debugf("todo repo")
	defaultExactMatch := map[string]bool{
		"id":   true,
		"_id":  true,
		"done": true, // bool match needs exact match, parram can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = Todo{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records Todos
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(todos []*Todo) ([]*Todo, error) {
	for _, todo := range todos {
		logger.Debugf("todo repo add: %+v", todo)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(Todos(todos))

	var records Todos
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(todos []*Todo) ([]*Todo, error) {
	logger.Debugf("todo repo update")
	rows, err := r.db.Save(Todos(todos))

	var records Todos
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("todo repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
