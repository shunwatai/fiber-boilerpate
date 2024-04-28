package document

import (
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/utils"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetId(t *testing.T) {
	document := &Document{
		MongoId: utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
		Id:      utils.ToPtr(helper.FlexInt(2)),
	}

	tests := []struct {
		name     string
		dbDriver string
		input    *Document
		want     string
	}{
		{name: "test Id", dbDriver: "postgres-or-mariadb-or-sqlite", input: document, want: "2"},
		{name: "test MongoId", dbDriver: "mongodb", input: document, want: "xxxx-xxxx-xxxx-xxxx"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg.LoadEnvVariables()
			cfg.Vpr.Set("database.engine", testCase.dbDriver)
			if err := cfg.Vpr.Unmarshal(cfg); err != nil {
				t.Logf("failed loading conf, err: %+v\n", err.Error())
			}

			got := document.GetId()
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
	documents := Documents{
		&Document{
			MongoId:   utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
			Id:        utils.ToPtr(helper.FlexInt(id)),
			FileSize:  1234,
			CreatedAt: customDatetime,
			UpdatedAt: customDatetime,
		},
	}

	tests := []struct {
		name  string
		input Documents
		want  []map[string]interface{}
	}{
		{name: "test StructToMap", input: documents, want: []map[string]interface{}{
			{"_id": "xxxx-xxxx-xxxx-xxxx", "id": float64(2), "created_at": timeJson, "updated_at": timeJson, "name": "", "file_path": "", "file_type": "", "file_size": float64(1234), "hash": "", "public": false, "user_id": nil, "user": nil},
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := documents.StructToMap()
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	documents := Documents{
		&Document{},
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "test get db tags", input: "db", want: []string{"id", "user_id", "name", "file_path", "file_type", "file_size", "hash", "public", "created_at", "updated_at"}},
		{name: "test get bson tags", input: "bson", want: []string{"_id", "id", "user_id", "name", "file_path", "file_type", "file_size", "hash", "public", "created_at", "updated_at"}},
		{name: "test get json tags", input: "json", want: []string{"_id", "id", "userId", "user", "name", "filePath", "fileType", "fileSize", "hash", "public", "createdAt", "updatedAt"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := documents.GetTags(testCase.input)
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
