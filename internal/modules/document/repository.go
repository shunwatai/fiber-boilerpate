package document

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

// cascadeFields for joining other module, see the example in internal/modules/document/repository.go
func cascadeFields(documents Documents) {
	// cascade user
}

func (r *Repository) Get(queries map[string]interface{}) ([]*Document, *helper.Pagination) {
	logger.Debugf("document repo")
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

	queries["columns"] = Document{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records Documents
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	//cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(documents []*Document) ([]*Document, error) {
	for _, document := range documents {
		logger.Debugf("document repo add: %+v", document)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(Documents(documents))

	var records Documents
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(documents []*Document) ([]*Document, error) {
	logger.Debugf("document repo update")
	rows, err := r.db.Save(Documents(documents))

	var records Documents
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("document repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
