package groupUser

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/user"
	"slices"

	//"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type GroupUser struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id        *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	GroupId   interface{}            `json:"groupId" db:"group_id" bson:"group_id,omitempty" validate:"omitempty,id_custom_validation"`
	UserId    interface{}            `json:"userId" db:"user_id" bson:"user_id,omitempty" validate:"omitempty,id_custom_validation"`
	User      *user.User             `json:"user"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type GroupUsers []*GroupUser

func (gu *GroupUser) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *gu.MongoId
	} else {
		return strconv.Itoa(int(*gu.Id))
	}
}

func (gu *GroupUser) GetGroupId() string {
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

func (gu *GroupUser) GetUserId() string {
	if cfg.DbConf.Driver == "mongodb" {
		userId, ok := gu.UserId.(string)
		if !ok {
			return ""
		}
		return userId
	} else {
		return strconv.Itoa(int(gu.UserId.(int64)))
	}
}

func (gus GroupUsers) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, gu := range gus {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(gu)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (gus GroupUsers) rowsToStruct(rows database.Rows) []*GroupUser {
	defer rows.Close()

	records := make([]*GroupUser, 0)
	for rows.Next() {
		var gu GroupUser
		err := rows.StructScan(&gu)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &gu)
	}

	return records
}

func (gus GroupUsers) GetTags(key string) []string {
	if len(gus) == 0 {
		return []string{}
	}

	return gus[0].getTags(key)
}

func (gus *GroupUsers) printValue() {
	for _, v := range *gus {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", v.GetId(), *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (gu GroupUser) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(gu)
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
