package log

import (
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/utils"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetId(t *testing.T) {
	log := &Log{
		MongoId: utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
		Id:      utils.ToPtr(helper.FlexInt(2)),
	}

	tests := []struct {
		name     string
		dbDriver string
		input    *Log
		want     string
	}{
		{name: "test Id", dbDriver: "postgres-or-mariadb-or-sqlite", input: log, want: "2"},
		{name: "test MongoId", dbDriver: "mongodb", input: log, want: "xxxx-xxxx-xxxx-xxxx"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg.LoadEnvVariables()
			cfg.Vpr.Set("database.engine", testCase.dbDriver)
			if err := cfg.Vpr.Unmarshal(cfg); err != nil {
				t.Logf("failed loading conf, err: %+v\n", err.Error())
			}

			got := log.GetId()
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	var id int64 = 2
	now := time.Now()
	customDatetime := &helper.CustomDatetime{&now, utils.ToPtr(time.RFC3339)}
	timeStr, _ := customDatetime.MarshalJSON()
	timeJson := strings.Replace(string(timeStr), "\"", "", -1)
	logs := Logs{
		&Log{
			MongoId:   utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
			Id:        utils.ToPtr(helper.FlexInt(id)),
			CreatedAt: customDatetime,
			UpdatedAt: customDatetime,
		},
	}

	tests := []struct {
		name  string
		input Logs
		want  []map[string]interface{}
	}{
		{name: "test StructToMap", input: logs, want: []map[string]interface{}{
			{"_id": "xxxx-xxxx-xxxx-xxxx", "id": float64(2), "created_at": timeJson, "updated_at": timeJson, "http_method": "", "duration": float64(0), "ip_address": "", "request_body": nil, "request_header": "", "response_body": nil, "route": "", "status": float64(0), "user_agent": "", "user_id": nil, "username": nil},
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := logs.StructToMap()
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	logs := Logs{
		&Log{},
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "test get db tags", input: "db", want: []string{"id", "user_id", "ip_address", "http_method", "route", "user_agent", "request_header", "request_body", "response_body", "status", "duration", "created_at", "updated_at"}},
		{name: "test get bson tags", input: "bson", want: []string{"_id", "id", "user_id", "ip_address", "http_method", "route", "user_agent", "request_header", "request_body", "response_body", "status", "duration", "created_at", "updated_at"}},
		{name: "test get json tags", input: "json", want: []string{"_id", "id", "userId", "username", "ipAddress", "httpMethod", "route", "userAgent", "requestHeader", "requestBody", "responseBody", "status", "duration", "createdAt", "updatedAt"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := logs.GetTags(testCase.input)
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
