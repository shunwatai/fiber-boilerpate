package resource

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"slices"

	//"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Resource struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"`
	Id        *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	Name      string                 `json:"name" db:"name" bson:"name,omitempty" validate:"required"`
	Order     int                    `json:"order" db:"order" bson:"order,omitempty"`
	Disabled  *bool                  `json:"disasbled" db:"disabled" bson:"disabled,omitempty" validate:"required,boolean"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type Resources []*Resource

func (rs Resources) GetNameMap() map[string]*Resource {
	nameMap := map[string]*Resource{}
	for _, r := range rs {
		nameMap[r.Name] = r
	}
	return nameMap
}

func (r *Resource) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *r.MongoId
	} else {
		return strconv.Itoa(int(*r.Id))
	}
}

func (rs Resources) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, r := range rs {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(r)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (rs Resources) rowsToStruct(rows database.Rows) []*Resource {
	defer rows.Close()

	records := make([]*Resource, 0)
	for rows.Next() {
		var r Resource
		err := rows.StructScan(&r)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &r)
	}

	return records
}

func (rs Resources) GetTags(key string) []string {
	if len(rs) == 0 {
		return []string{}
	}

	return rs[0].getTags(key)
}

func (rs *Resources) printValue() {
	for _, v := range *rs {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (r Resource) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(r)
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
