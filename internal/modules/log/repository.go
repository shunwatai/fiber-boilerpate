package log

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
func cascadeFields(logs Logs) {
	// cascade user
}

func (r *Repository) Get(queries map[string]interface{}) ([]*Log, *helper.Pagination) {
	logger.Debugf("log repo")
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

	queries["columns"] = Log{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records Logs
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	records.setUsername()
	//cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(logs []*Log) ([]*Log, error) {
	for _, log := range logs {
		logger.Debugf("log repo add: %+v", log)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(Logs(logs))

	var records Logs
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(logs []*Log) ([]*Log, error) {
	logger.Debugf("log repo update")
	rows, err := r.db.Save(Logs(logs))

	var records Logs
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("log repo delete\n")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
