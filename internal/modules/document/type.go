package document

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Document struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"`
	Id        *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	UserId    interface{}            `json:"userId" db:"user_id" bson:"user_id,omitempty" validate:"omitempty,id_custom_validation"`
	User      *user.User             `json:"user"`
	Name      string                 `json:"name" db:"name" bson:"name,omitempty" example:"test.jpg"`
	FilePath  string                 `json:"filePath" db:"file_path" bson:"file_path,omitempty" example:"upload/xx/202210041710-test.jpg"`
	FileType  string                 `json:"fileType" db:"file_type" bson:"file_type,omitempty" default:"jpg"`
	FileSize  int64                  `json:"fileSize" db:"file_size" bson:"file_size,omitempty" default:"342424"`
	Hash      string                 `json:"hash" db:"hash" bson:"hash"`
	Public    bool                   `json:"public" db:"public" bson:"public,omitempty" validate:"boolean"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type Documents []*Document

func (doc *Document) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *doc.MongoId
	} else {
		return strconv.Itoa(int(*doc.Id))
	}
}

//func (doc *Document) GetUserId() string {
//	if cfg.DbConf.Driver == "mongodb" {
//		userId, ok := doc.UserId.(string)
//		if !ok {
//			return ""
//		}
//		return userId
//	} else {
//		return strconv.Itoa(int(doc.UserId.(int64)))
//	}
//}

func (docs Documents) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, doc := range docs {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(doc)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (docs Documents) rowsToStruct(rows database.Rows) []*Document {
	defer rows.Close()

	records := make([]*Document, 0)
	for rows.Next() {
		var doc Document
		err := rows.StructScan(&doc)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &doc)
	}

	return records
}

func (docs Documents) GetTags(key string) []string {
	if len(docs) == 0 {
		return []string{}
	}

	return docs[0].getTags(key)
}

func (docs *Documents) printValue() {
	for _, v := range *docs {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (doc Document) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(doc)
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
