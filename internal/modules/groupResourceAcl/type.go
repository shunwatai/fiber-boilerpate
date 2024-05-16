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
	MongoId          *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id               *int64                 `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	GroupId          interface{}            `json:"groupId" db:"group_id" bson:"group_id,omitempty" validate:"omitempty,id_custom_validation"`
	GroupName        *string                `json:"groupName" db:"group_name" bson:"group_name,omitempty"`
	ResourceId       interface{}            `json:"resourceId" db:"resource_id" bson:"resource_id,omitempty" validate:"omitempty,id_custom_validation"`
	ResourceName     *string                `json:"resourceName" db:"resource_name" bson:"resource_name,omitempty"`
	PermissionTypeId interface{}            `json:"permissionTypeId" db:"permission_type_id" bson:"permission_type_id,omitempty" validate:"omitempty,id_custom_validation"`
	PermissionType   *string                `json:"permissionType" db:"permission_type" bson:"permission_type,omitempty"`
	CreatedAt        *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt        *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type GroupResourceAcls []*GroupResourceAcl

func (gra *GroupResourceAcl) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *gra.MongoId
	} else {
		return strconv.Itoa(int(*gra.Id))
	}
}

func (gu *GroupResourceAcl) GetGroupId() string {
	if cfg.DbConf.Driver == "mongodb" {
		groupId, ok := gu.GroupId.(string)
		if !ok {
			return ""
		}
		return groupId
	} else {
		return strconv.Itoa(int(gu.GroupId.(int64)))
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
		} else {
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
