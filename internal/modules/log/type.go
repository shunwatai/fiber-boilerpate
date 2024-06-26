package log

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/groupUser"
	"golang-api-starter/internal/modules/user"
	"slices"

	//"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Log struct {
	MongoId       *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"`
	Id            *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	UserId        interface{}            `json:"userId" db:"user_id" bson:"user_id" example:"1"`
	Username      *string                `json:"username"`
	IpAddress     string                 `json:"ipAddress" db:"ip_address" bson:"ip_address" example:"29.23.43.23"`
	HttpMethod    string                 `json:"httpMethod" db:"http_method" bson:"http_method" example:"GET"`
	Route         string                 `json:"route" db:"route" bson:"route" example:"/api/users"`
	UserAgent     string                 `json:"userAgent" db:"user_agent" bson:"user_agent" example:"postman"`
	RequestHeader string                 `json:"requestHeader" db:"request_header" bson:"request_header"`
	RequestBody   *string                `json:"requestBody" db:"request_body" bson:"request_body"`
	ResponseBody  *string                `json:"responseBody" db:"response_body" bson:"response_body"`
	Status        int64                  `json:"status" db:"status" bson:"status" example:"200"`
	Duration      int64                  `json:"duration" db:"duration" bson:"duration"`
	CreatedAt     *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt     *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type Logs []*Log

func (lg *Log) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *lg.MongoId
	} else {
		return strconv.Itoa(int(*lg.Id))
	}
}

func (lg *Log) GetUserId() string {
	if cfg.DbConf.Driver == "mongodb" {
		userId, ok := lg.UserId.(string)
		if !ok {
			return ""
		}
		return userId
	} else {
		return strconv.Itoa(int(lg.UserId.(int64)))
	}
}

func (lgs Logs) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, lg := range lgs {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(lg)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (lgs Logs) rowsToStruct(rows database.Rows) []*Log {
	defer rows.Close()

	records := make([]*Log, 0)
	for rows.Next() {
		var lg Log
		err := rows.StructScan(&lg)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &lg)
	}

	return records
}

func (lgs Logs) GetTags(key ...string) []string {
	if len(lgs) == 0 {
		return []string{}
	}

	return lgs[0].getTags(key...)
}

func (lgs *Logs) printValue() {
	for _, v := range *lgs {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (lg Log) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(lg)
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

func (logs Logs) setUsername(){
	var (
		userIds []string
		userId  string
	)
	// get all userIds
	for _, log := range logs {
		if log.UserId == nil {
			continue
		}

		userId = log.GetUserId()
		userIds = append(userIds, userId)
	}

	// if no userIds, do nothing and return
	if len(userIds) > 0 {
		users := []*groupUser.User{}

		// get users by userIds
		condition := database.GetIdsMapCondition(nil, userIds)
		users, _ = user.Srvc.Get(condition)
		// get the map[userId]user
		userMap := user.Repo.GetIdMap(users)

		for _, log := range logs {
			if log.UserId == nil {
				continue
			}
			user := &groupUser.User{}
			// take out the user by userId in map and assign
			user = userMap[log.GetUserId()]
			log.Username = &user.Name
		}
	}
}
