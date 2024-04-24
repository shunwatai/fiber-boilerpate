package todoDocument

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/document"
	"slices"

	//"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type TodoDocument struct {
	MongoId    *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id         *int64                 `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	TodoId     interface{}            `json:"todoId" db:"todo_id" bson:"todo_id,omitempty" validate:"omitempty,id_custom_validation"`
	DocumentId interface{}            `json:"documentId" db:"document_id" bson:"document_id,omitempty" validate:"omitempty,id_custom_validation"`
	Document   *document.Document     `json:"document"`
	CreatedAt  *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt  *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type TodoDocuments []*TodoDocument

func (td *TodoDocument) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *td.MongoId
	} else {
		return strconv.Itoa(int(*td.Id))
	}
}

func (td *TodoDocument) GetTodoId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return td.TodoId.(string)
	} else {
		return strconv.Itoa(int(td.TodoId.(int64)))
	}
}

//func (td *TodoDocument) GetUserId() string {
//	if cfg.DbConf.Driver == "mongodb" {
//		userId, ok := td.UserId.(string)
//		if !ok {
//			return ""
//		}
//		return userId
//	} else {
//		return strconv.Itoa(int(td.UserId.(int64)))
//	}
//}

func (td *TodoDocument) GetDocumentId() string {
	if cfg.DbConf.Driver == "mongodb" {
		userId, ok := td.DocumentId.(string)
		if !ok {
			return ""
		}
		return userId
	} else {
		return strconv.Itoa(int(td.DocumentId.(int64)))
	}
}

func (tds TodoDocuments) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, td := range tds {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(td)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (tds TodoDocuments) rowsToStruct(rows database.Rows) []*TodoDocument {
	defer rows.Close()

	records := make([]*TodoDocument, 0)
	for rows.Next() {
		var td TodoDocument
		err := rows.StructScan(&td)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &td)
	}

	return records
}

func (tds TodoDocuments) GetTags(key string) []string {
	if len(tds) == 0 {
		return []string{}
	}

	return tds[0].getTags(key)
}

func (tds *TodoDocuments) printValue() {
	for _, v := range *tds {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (td TodoDocument) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(td)
	for i := 0; i < val.Type().NumField(); i++ {
		t := val.Type().Field(i)
		fieldName := t.Name

		switch jsonTag := t.Tag.Get(tag); jsonTag {
		case "-":
		case "":
			// fmt.Println(fieldName)
		default:
			parts := strings.Split(jsonTag, ",")
			name := parts[0]
			if name == "" {
				name = fieldName
			}
			// fmt.Println(name)
			if !slices.Contains(*database.IgnrCols, name) {
				cols = append(cols, name)
			}
		}
	}
	return cols
}
