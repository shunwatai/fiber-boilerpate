package permissionType

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

type PermissionType struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"`
	Id        *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	Name      string                 `json:"name" db:"name" bson:"name,omitempty" validate:"required"`
	Order     int                    `json:"order" db:"order" bson:"order,omitempty"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type PermissionTypes []*PermissionType

func (pts PermissionTypes) GetNameMap() map[string]*PermissionType {
	nameMap := map[string]*PermissionType{}
	for _, pt := range pts {
		nameMap[pt.Name] = pt
	}
	return nameMap
}

func (pt *PermissionType) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *pt.MongoId
	} else {
		return strconv.Itoa(int(*pt.Id))
	}
}

func (pts PermissionTypes) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, pt := range pts {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(pt)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (pts PermissionTypes) rowsToStruct(rows database.Rows) []*PermissionType {
	defer rows.Close()

	records := make([]*PermissionType, 0)
	for rows.Next() {
		var pt PermissionType
		err := rows.StructScan(&pt)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &pt)
	}

	return records
}

func (pts PermissionTypes) GetTags(key ...string) []string {
	if len(pts) == 0 {
		return []string{}
	}

	return pts[0].getTags(key...)
}

func (pts *PermissionTypes) printValue() {
	for _, v := range *pts {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (pt PermissionType) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(pt)
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
