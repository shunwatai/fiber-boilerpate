package groupResourceAcl

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

type GroupResourceAcl struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id        *int64                 `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	//UserId    interface{}            `json:"userId" db:"user_id" bson:"user_id,omitempty" validate:"omitempty,id_custom_validation"`
	//User      *user.User             `json:"user"`
	Col1      string                 `json:"col1" db:"col_1" bson:"col_1,omitempty" validate:"required"`
	Col2      *bool                  `json:"col2" db:"col_2" bson:"col_2,omitempty" validate:"required,boolean"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type GroupResourceAcls []*GroupResourceAcl

func (gra *GroupResourceAcl) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *gra.MongoId
	} else {
		return strconv.Itoa(int(*gra.Id))
	}
}

//func (gra *GroupResourceAcl) GetUserId() string {
//	if cfg.DbConf.Driver == "mongodb" {
//		userId, ok := gra.UserId.(string)
//		if !ok {
//			return ""
//		}
//		return userId
//	} else {
//		return strconv.Itoa(int(gra.UserId.(int64)))
//	}
//}

func (gras GroupResourceAcls) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, gra := range gras {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(gra)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (gras GroupResourceAcls) rowsToStruct(rows database.Rows) []*GroupResourceAcl {
	defer rows.Close()

	records := make([]*GroupResourceAcl, 0)
	for rows.Next() {
		var gra GroupResourceAcl
		err := rows.StructScan(&gra)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &gra)
	}

	return records
}

func (gras GroupResourceAcls) GetTags(key string) []string {
	if len(gras) == 0 {
		return []string{}
	}

	return gras[0].getTags(key)
}

func (gras *GroupResourceAcls) printValue() {
	for _, v := range *gras {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		}else{
      logger.Debugf("new --> v: %+v\n", *v)
    }
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (gra GroupResourceAcl) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(gra)
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
