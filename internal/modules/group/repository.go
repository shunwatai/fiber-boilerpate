package group

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/groupResourceAcl"
	"golang-api-starter/internal/modules/groupUser"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db       database.IDatabase
	UserRepo groupUser.IUserRepository
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetIdMap(groups groupUser.Groups) map[string]*groupUser.Group {
	groupMap := map[string]*groupUser.Group{}
	for _, group := range groups {
		groupMap[group.GetId()] = group
	}
	return groupMap
}

// cascadeFields for joining other module, see the example in internal/modules/todo/repository.go
func cascadeFields(groups groupUser.Groups) {
	if len(groups) == 0 {
		return
	}

	var groupIds []string
	// get all gruopId
	for _, group := range groups {
		groupId := group.GetId()
		groupIds = append(groupIds, groupId)
	}

	condition := database.GetIdsMapCondition(utils.ToPtr("group_id"), groupIds)

	// get groupUsers by groupId
	groupUsers, _ := groupUser.Repo.Get(condition)
	groupUsersMap := groupUser.Repo.GetGroupIdMap(groupUsers)

	// get groupResourceAcls by groupId
	groupResourceAcls, _ := groupResourceAcl.Repo.Get(condition)
	groupAclsMap := groupResourceAcl.Repo.GetGroupIdMap(groupResourceAcls)

	// map users & permission into group
	for _, group := range groups {
		// if no users, assign empty slice for response json "users": [] instead of "users": null
		group.Users = []*groupUser.User{}
		// take out the groupUsers by groupId in map and assign
		gus, haveUsers := groupUsersMap[group.GetId()]

		if haveUsers {
			for _, gu := range gus {
				gu.User.Groups = nil
				group.Users = append(group.Users, gu.User)
			}
		}

		// if no permissions, assign empty slice for response json "permissions": [] instead of "permissions": null
		group.Permissions = []*groupResourceAcl.GroupResourceAcl{}
		// take out the groupUsers by groupId in map and assign
		gas, haveAcls := groupAclsMap[group.GetId()]

		if haveAcls {
			for _, ga := range gas {
				group.Permissions = append(group.Permissions, ga)
			}
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*groupUser.Group, *helper.Pagination) {
	logger.Debugf("group repo get")
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

	queries["columns"] = groupUser.Groups{{}}.GetTags()
	rows, pagination := r.db.Select(queries)

	var records groupUser.Groups
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	// records.printValue()

	return records, pagination
}

func (r *Repository) Create(groups []*groupUser.Group) ([]*groupUser.Group, error) {
	for _, group := range groups {
		logger.Debugf("group repo add: %+v", group)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(groupUser.Groups(groups))

	var records groupUser.Groups
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	records.PrintValue()

	return records, err
}

func (r *Repository) Update(groups []*groupUser.Group) ([]*groupUser.Group, error) {
	logger.Debugf("group repo update")
	rows, err := r.db.Save(groupUser.Groups(groups))

	var records groupUser.Groups
	if rows != nil {
		records = records.RowsToStruct(rows)
	}
	records.PrintValue()
	// cascadeFields(records)

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("group repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
