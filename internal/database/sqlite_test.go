package database

import (
	"database/sql"
	"fmt"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func setupSqliteTestTable(t *testing.T) func(t *testing.T) {
	t.Logf("setup sqlite test table\n")
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "sqlite")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	zlog.NewZlog()
	var testDb = GetDatabase("todos_test",nil)
	testDb.Connect()

	// create test table
	testDb.RawQuery(`CREATE TABLE IF NOT EXISTS "todos_test" ( "id"	INTEGER NOT NULL, "task"	TEXT, "done"	NUMERIC, "created_at"	DATETIME, "updated_at"	DATETIME, PRIMARY KEY("id" AUTOINCREMENT));`)

	// insert dummy data
	testDb.RawQuery(`INSERT INTO ` + "todos_test" + ` (id,task,done,created_at,updated_at) VALUES 
		(2,'want sleep',0,'2023-11-03','2023-11-03'),
		(3,'stop code',0,'2023-11-03','2023-11-03'),
		(13,'take shower',0,'2023-11-03','2023-11-03'),
		(15,'want sleep',0,'2023-11-03','2023-11-03'),
		(4,'want sleep',0,'2023-11-03','2023-11-03'),
		(44,'want sleep',0,'2023-11-03','2023-11-03'),
		(41,'want sleep',0,'2023-11-03','2023-11-03')
	`)

	return func(t *testing.T) {
		t.Log("teardown sqlite test table")
		// testDb.RawQuery("DROP TABLE IF EXISTS `todos_test`;")
	}
}

type sqliteTests struct {
	name  string
	input map[string]interface{}
	want1 string
	want2 map[string]interface{}
}

func TestSqliteConstructSelectStmtFromQuerystring(t *testing.T) {
	dir := t.TempDir()
	sqlitedbFile := filepath.Join(dir, "test.db")
	connStr := fmt.Sprintf("%s?_auth&_auth_user=user&_auth_pass=password&_auth_crypt=sha1&parseTime=true", sqlitedbFile)
	log.Printf("tmp dir::: %+v\n", connStr)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			return
		}
	}()
	if err = db.Ping(); err != nil {
		t.Fatal("failed to connect to sqlite....")
	}
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "sqlite")
	cfg.Vpr.Set("database.sqlite.user", "user")
	cfg.Vpr.Set("database.sqlite.pass", "password")
	cfg.Vpr.Set("database.sqlite.database", connStr) // put the connStr as "database" when running test
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}

	teardownTest := setupSqliteTestTable(t)
	defer teardownTest(t)

	var tableName = "todos_test"
	var testDb = GetDatabase(tableName,nil)
	testDb.Connect()
	tests := []sqliteTests{
		{
			name:  "get by ID",
			input: map[string]interface{}{"id": 2, "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE id=:id ORDER BY id desc LIMIT 1 OFFSET 0`,
			want2: map[string]interface{}{"id": 2},
		},
		{
			name:  "get by IDs",
			input: map[string]interface{}{"id": []string{"2", "3"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE id IN (:id1,:id2) ORDER BY id desc LIMIT 2 OFFSET 0`,
			want2: map[string]interface{}{"id1": "2", "id2": "3"},
		},
		{
			name:  "get keyword by ILIKE",
			input: map[string]interface{}{"task": "show", "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE task like :task ORDER BY id desc LIMIT 1 OFFSET 0`,
			want2: map[string]interface{}{"task": "%show%"},
		},
		{
			name:  "get keywords by ~~ ANY(xx)",
			input: map[string]interface{}{"task": []string{"show", "stop"}, "page": "1", "items": "5", "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE (lower(task) like :task1 or lower(task) like :task2)  ORDER BY id desc LIMIT 5 OFFSET 0`,
			want2: map[string]interface{}{"task1": "%show%", "task2": "%stop%"},
		},
		{
			name:  "get records by keyword that matches in given ids",
			input: map[string]interface{}{"task": "wan", "id": []string{"13", "15"}, "page": "1", "items": "5", "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE id IN (:id1,:id2) AND task like :task ORDER BY id desc LIMIT 5 OFFSET 0`,
			want2: map[string]interface{}{"task": "%wan%", "id1": "13", "id2": "15"},
		},
		{
			name:  "get records by date range",
			input: map[string]interface{}{"withDateFilter": true, "created_at": "2023-01-01.2023-12-31", "page": "1", "items": "5", "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
			want1: `SELECT * FROM todos_test WHERE created_at >= :created_atFrom AND created_at <= :created_atTo ORDER BY id desc LIMIT 5 OFFSET 0`,
			want2: map[string]interface{}{"created_atFrom": "2023-01-01", "created_atTo": "2023-12-31"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got1, _, got2 := testDb.constructSelectStmtFromQuerystring(testCase.input)

			if eq := reflect.DeepEqual(strings.Fields(testCase.want1), strings.Fields(got1)); !eq {
				t.Errorf("got %q \nwant %q", strings.Fields(got1), strings.Fields(testCase.want1))
			}

			if eq := reflect.DeepEqual(testCase.want2, got2); !eq {
				t.Errorf("got %+v \nwant %+v", got2, testCase.want2)
			}
		})
	}
}
