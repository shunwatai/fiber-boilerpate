package todo

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/document"
	"golang-api-starter/internal/modules/groupUser"
	"log"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Todo struct {
	MongoId       *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id            *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	UserId        interface{}            `json:"userId" db:"user_id" bson:"user_id,omitempty" validate:"omitempty,id_custom_validation"`
	User          *groupUser.User             `json:"user"`
	TodoDocuments interface{}            `json:"-"`
	Documents     []*document.Document   `json:"documents"`
	Task          string                 `json:"task" db:"task" bson:"task,omitempty" validate:"required"`
	Done          *bool                  `json:"done" db:"done" bson:"done,omitempty" validate:"required,boolean"`
	CreatedAt     *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt     *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
	// CreatedAt *string `db:"created_at" json:"createdAt,omitempty"`
	// UpdatedAt *string `db:"updated_at" json:"updatedAt,omitempty"`
}

type Todos []*Todo

func (todo *Todo) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *todo.MongoId
	} else {
		return strconv.Itoa(int(*todo.Id))
	}
}

func (todo *Todo) GetUserId() string {
	if cfg.DbConf.Driver == "mongodb" {
		userId, ok := todo.UserId.(string)
		if !ok {
			return ""
		}
		return userId
	} else {
		return strconv.Itoa(int(todo.UserId.(int64)))
	}
}

func (todos Todos) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, todo := range todos {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(todo)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (todos Todos) rowsToStruct(rows database.Rows) []*Todo {
	defer rows.Close()

	records := make([]*Todo, 0)
	for rows.Next() {
		var todo Todo
		err := rows.StructScan(&todo)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &todo)
	}

	return records
}

func (todos Todos) GetTags(key ...string) []string {
	if len(todos) == 0 {
		return []string{}
	}

	return todos[0].getTags(key...)
}

func (todos *Todos) printValue() {
	for _, v := range *todos {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (todo Todo) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(todo)
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
