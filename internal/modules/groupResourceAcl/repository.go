package groupResourceAcl

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
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
func cascadeFields(groupResourceAcls GroupResourceAcls) {
  if len(groupResourceAcls) == 0 {
    return
  }
	// cascade user
}

func (r *Repository) Get(queries map[string]interface{}) ([]*GroupResourceAcl, *helper.Pagination) {
	logger.Debugf("groupResourceAcl repo get")
	defaultExactMatch := map[string]bool{
		"id":   true,
		"_id":  true,
		//"done": true, // bool match needs exact match, param can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = GroupResourceAcl{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records GroupResourceAcls
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	//cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(groupResourceAcls []*GroupResourceAcl) ([]*GroupResourceAcl, error) {
	for _, groupResourceAcl := range groupResourceAcls {
		logger.Debugf("groupResourceAcl repo add: %+v", groupResourceAcl)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(GroupResourceAcls(groupResourceAcls))

	var records GroupResourceAcls
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(groupResourceAcls []*GroupResourceAcl) ([]*GroupResourceAcl, error) {
	logger.Debugf("groupResourceAcl repo update")
	rows, err := r.db.Save(GroupResourceAcls(groupResourceAcls))

	var records GroupResourceAcls
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("groupResourceAcl repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
