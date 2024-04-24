package todoDocument

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/document"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db database.IDatabase
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db}
}

// cascadeFields for joining other module, see the example in internal/modules/todo/repository.go
func cascadeFields(todoDocuments TodoDocuments) {
	if len(todoDocuments) == 0 {
		return
	}
	// cascade documents
	var (
		documentIds []string
		documentId  string
	)
	// get all userIds
	for _, todoDocument := range todoDocuments {
		if todoDocument.DocumentId == nil {
			continue
		}

		documentId = todoDocument.GetDocumentId()
		documentIds = append(documentIds, documentId)
	}

	if len(documentIds) > 0 {
		documents := []*document.Document{}

		// get documents by documentsIds
		condition := database.GetIdsMapCondition(nil, documentIds)
		documents, _ = document.Srvc.Get(condition)
		// get the map[documentId]document
		documentMap := document.Srvc.GetIdMap(documents)

		for _, todoDocument := range todoDocuments {
			if todoDocument.DocumentId == nil {
				continue
			}
			document := &document.Document{}
			// take out the document by documentId in map and assign
			document = documentMap[todoDocument.GetDocumentId()]
			todoDocument.Document = document
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*TodoDocument, *helper.Pagination) {
	logger.Debugf("todoDocument repo")
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

	queries["columns"] = TodoDocument{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records TodoDocuments
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(todoDocuments []*TodoDocument) ([]*TodoDocument, error) {
	for _, todoDocument := range todoDocuments {
		logger.Debugf("todoDocument repo add: %+v", todoDocument)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(TodoDocuments(todoDocuments))

	var records TodoDocuments
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(todoDocuments []*TodoDocument) ([]*TodoDocument, error) {
	logger.Debugf("todoDocument repo update")
	rows, err := r.db.Save(TodoDocuments(todoDocuments))

	var records TodoDocuments
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("todoDocument repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
