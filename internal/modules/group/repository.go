package group

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
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
	groups.SetUsers()
	groups.SetPermissions()
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
