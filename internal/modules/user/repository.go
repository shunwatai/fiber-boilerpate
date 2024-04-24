package user

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db database.IDatabase
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*User, *helper.Pagination) {
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

	queries["columns"] = User{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records Users
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	return records, pagination
}

func (r *Repository) Create(users []*User) ([]*User, error) {
	logger.Debugf("user repo create")
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(Users(users))

	var records Users
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(users []*User) ([]*User, error) {
	logger.Debugf("user repo update")
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(Users(users))

	var records Users
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

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
