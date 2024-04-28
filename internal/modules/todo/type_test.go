package todo

import (
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/utils"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetId(t *testing.T) {
	todo := &Todo{
		MongoId: utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
		Id:      utils.ToPtr(helper.FlexInt(2)),
	}

	tests := []struct {
		name     string
		dbDriver string
		input    *Todo
		want     string
	}{
		{name: "test Id", dbDriver: "postgres-or-mariadb-or-sqlite", input: todo, want: "2"},
		{name: "test MongoId", dbDriver: "mongodb", input: todo, want: "xxxx-xxxx-xxxx-xxxx"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg.LoadEnvVariables()
			cfg.Vpr.Set("database.engine", testCase.dbDriver)
			if err := cfg.Vpr.Unmarshal(cfg); err != nil {
				t.Logf("failed loading conf, err: %+v\n", err.Error())
			}

			got := todo.GetId()
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestGetUserId(t *testing.T) {
	tests := []struct {
		name     string
		dbDriver string
		input    *Todo
		want     string
	}{
		{name: "test userId", dbDriver: "postgres-or-mariadb-or-sqlite", input: &Todo{UserId: int64(2)}, want: "2"},
		{name: "test mongo userId", dbDriver: "mongodb", input: &Todo{UserId: "xxxx-xxxx-xxxx-xxxx"}, want: "xxxx-xxxx-xxxx-xxxx"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg.LoadEnvVariables()
			cfg.Vpr.Set("database.engine", testCase.dbDriver)
			if err := cfg.Vpr.Unmarshal(cfg); err != nil {
				t.Logf("failed loading conf, err: %+v\n", err.Error())
			}

			got := testCase.input.GetUserId()
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
	todos := Todos{
		&Todo{
			MongoId:   utils.ToPtr("xxxx-xxxx-xxxx-xxxx"),
			Id:        utils.ToPtr(helper.FlexInt(id)),
			CreatedAt: customDatetime,
			UpdatedAt: customDatetime,
		},
	}

	tests := []struct {
		name  string
		input Todos
		want  []map[string]interface{}
	}{
		{name: "test StructToMap", input: todos, want: []map[string]interface{}{
			{"_id": "xxxx-xxxx-xxxx-xxxx", "id": float64(2), "created_at": timeJson, "updated_at": timeJson, "documents": nil, "done": nil, "user": nil, "user_id": nil, "task": ""},
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := todos.StructToMap()
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	todos := Todos{
		&Todo{},
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "test get db tags", input: "db", want: []string{"id", "user_id", "task", "done", "created_at", "updated_at"}},
		{name: "test get bson tags", input: "bson", want: []string{"_id", "id", "user_id", "task", "done", "created_at", "updated_at"}},
		{name: "test get json tags", input: "json", want: []string{"_id", "id", "userId", "user", "documents", "task", "done", "createdAt", "updatedAt"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := todos.GetTags(testCase.input)
			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
