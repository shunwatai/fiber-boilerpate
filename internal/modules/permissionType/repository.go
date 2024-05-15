package permissionType

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
func cascadeFields(permissionTypes PermissionTypes) {
  if len(permissionTypes) == 0 {
    return
  }
	// cascade user
}

func (r *Repository) Get(queries map[string]interface{}) ([]*PermissionType, *helper.Pagination) {
	logger.Debugf("permissionType repo get")
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

	queries["columns"] = PermissionType{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records PermissionTypes
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	//cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(permissionTypes []*PermissionType) ([]*PermissionType, error) {
	for _, permissionType := range permissionTypes {
		logger.Debugf("permissionType repo add: %+v", permissionType)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(PermissionTypes(permissionTypes))

	var records PermissionTypes
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(permissionTypes []*PermissionType) ([]*PermissionType, error) {
	logger.Debugf("permissionType repo update")
	rows, err := r.db.Save(PermissionTypes(permissionTypes))

	var records PermissionTypes
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("permissionType repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
