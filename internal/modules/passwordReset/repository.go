package passwordReset

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
func cascadeFields(passwordResets PasswordResets) {
	if len(passwordResets) == 0 {
		return
	}
	// cascade user
}

func (r *Repository) Get(queries map[string]interface{}) ([]*PasswordReset, *helper.Pagination) {
	logger.Debugf("passwordReset repo get")
	defaultExactMatch := map[string]bool{
		"id":      true,
		"_id":     true,
		"is_used": true, // bool match needs exact match, param can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = PasswordReset{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records PasswordResets
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	//cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(passwordResets []*PasswordReset) ([]*PasswordReset, error) {
	for _, passwordReset := range passwordResets {
		logger.Debugf("passwordReset repo add: %+v", passwordReset)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(PasswordResets(passwordResets))

	var records PasswordResets
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(passwordResets []*PasswordReset) ([]*PasswordReset, error) {
	logger.Debugf("passwordReset repo update")
	rows, err := r.db.Save(PasswordResets(passwordResets))

	var records PasswordResets
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("passwordReset repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
