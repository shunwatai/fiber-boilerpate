package user

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/groupUser"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db        database.IDatabase
	GroupRepo groupUser.IGroupRepository
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetIdMap(users groupUser.Users) map[string]*groupUser.User {
	userMap := map[string]*groupUser.User{}
	sanitise(users)
	for _, user := range users {
		userMap[user.GetId()] = user
	}
	return userMap
}

// cascadeFields for joining other module, see the example in internal/modules/todo/repository.go
func cascadeFields(users groupUser.Users) {
	if len(users) == 0 {
		return
	}

	var userIds []string
	// get all gruopId
	for _, user := range users {
		userId := user.GetId()
		userIds = append(userIds, userId)
	}

	condition := database.GetIdsMapCondition(utils.ToPtr("user_id"), userIds)

	// get groupUsers by groupId
	groupUsers, _ := groupUser.Repo.Get(condition)
	groupUsersMap := groupUser.Repo.GetUserIdMap(groupUsers)

	// map users & permission into group
	for _, user := range users {
		// if no users, assign empty slice for response json "users": [] instead of "users": null
		user.Groups = []*groupUser.Group{}
		// take out the groupUsers by groupId in map and assign
		gus, haveUsers := groupUsersMap[user.GetId()]

		if haveUsers {
			for _, gu := range gus {
				gu.Group.Users = nil
				user.Groups = append(user.Groups, gu.Group)
			}
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*groupUser.User, *helper.Pagination) {
	logger.Debugf("user repo get")
	defaultExactMatch := map[string]bool{
		"id":       true,
		"_id":      true,
		"disabled": true, // bool match needs exact match, parram can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = groupUser.Users{{}}.GetTags()
	rows, pagination := r.db.Select(queries)

	var records groupUser.Users
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	// records.PrintValue()

	return records, pagination
}

func (r *Repository) Create(users []*groupUser.User) ([]*groupUser.User, error) {
	logger.Debugf("user repo create")
	*database.IgnrCols = append(*database.IgnrCols,"search")
	database.SetIgnoredCols(*database.IgnrCols...)
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(groupUser.Users(users))

	var records groupUser.Users
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	records.PrintValue()

	return records, err
}

func (r *Repository) Update(users []*groupUser.User) ([]*groupUser.User, error) {
	logger.Debugf("user repo update")
	*database.IgnrCols = append(*database.IgnrCols,"search")
	database.SetIgnoredCols(*database.IgnrCols...)
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(groupUser.Users(users))

	var records groupUser.Users
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	records.PrintValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("user repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
