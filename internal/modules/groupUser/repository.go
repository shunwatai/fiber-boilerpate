package groupUser

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/user"

	//"golang-api-starter/internal/modules/user"
	"golang.org/x/exp/maps"
)

type Repository struct {
	db database.IDatabase
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db}
}

// cascadeFields for joining other module, see the example in internal/modules/todo/repository.go
func cascadeFields(groupUsers GroupUsers) {
	if len(groupUsers) == 0 {
		return
	}

	// cascade user
	// get users by userIds
	var (
		userIds []string
		userId  string
	)
	// get all userIds
	for _, groupUser := range groupUsers {
		userId = groupUser.GetUserId()
		userIds = append(userIds, userId)
	}

	if len(userIds) > 0 {
		// get documents by documentsIds
		condition := database.GetIdsMapCondition(utils.ToPtr("user_id"), userIds)
		users, _ := user.Srvc.Get(condition)
		// get the map[userId]User
		userMap := user.Srvc.GetIdMap(users)

		for _, groupUser := range groupUsers {
			groupUser.User = new(user.User)
			// take out the document by documentId in map and assign
			user := userMap[groupUser.GetUserId()]
			groupUser.User = user
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*GroupUser, *helper.Pagination) {
	logger.Debugf("groupUser repo get")
	defaultExactMatch := map[string]bool{
		"id":  true,
		"_id": true,
		//"done": true, // bool match needs exact match, param can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = GroupUser{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(groupUsers []*GroupUser) ([]*GroupUser, error) {
	for _, groupUser := range groupUsers {
		logger.Debugf("groupUser repo add: %+v", groupUser)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(GroupUsers(groupUsers))

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(groupUsers []*GroupUser) ([]*GroupUser, error) {
	logger.Debugf("groupUser repo update")
	rows, err := r.db.Save(GroupUsers(groupUsers))

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("groupUser repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
