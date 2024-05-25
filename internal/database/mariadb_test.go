//go:build integration

package database

import (
	"context"
	"database/sql"
	"fmt"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/dhui/dktest"
)

var (
	mariadbOps = dktest.Options{
		PortRequired: true,
		ReadyFunc:    mariadbReady,
		Env:          map[string]string{"MYSQL_ROOT_PASSWORD": "root", "MYSQL_DATABASE": "public"},
	}

	mariadbReady = func(ctx context.Context, c dktest.ContainerInfo) bool {
		ip, port, err := c.FirstPort()
		if err != nil {
			return false
		}

		connStr := fmt.Sprintf("root:root@tcp(%s:%s)/public?charset=utf8&parseTime=True&loc=Local", ip, port)
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			return false
		}

		defer db.Close()
		return db.PingContext(ctx) == nil
	}
)

func setupMariadbTestTable(t *testing.T) func(t *testing.T) {
	t.Logf("setup mariadb test table\n")
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "mariadb")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	zlog.NewZlog()
	var testDb = GetDatabase("", nil)

	// create test table
	testDb.RawQuery("CREATE TABLE IF NOT EXISTS  `todos_test` ( `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY, `task` varchar(255) NOT NULL, `done` tinyint(1) NOT NULL DEFAULT '0', `created_at` datetime NOT NULL DEFAULT current_timestamp, `updated_at` datetime NOT NULL DEFAULT current_timestamp ON UPDATE CURRENT_TIMESTAMP);")

	// insert dummy data
	testDb.RawQuery(`INSERT INTO .todos_test (id,task,done,created_at,updated_at) VALUES 
		(2,'want sleep',0,'2023-11-03','2023-11-03'),
		(3,'stop code',0,'2023-11-03','2023-11-03'),
		(13,'take shower',0,'2023-11-03','2023-11-03'),
		(15,'want sleep',0,'2023-11-03','2023-11-03'),
		(4,'want sleep',0,'2023-11-03','2023-11-03'),
		(44,'want sleep',0,'2023-11-03','2023-11-03'),
		(41,'want sleep',0,'2023-11-03','2023-11-03')
	`)

	return func(t *testing.T) {
		t.Log("teardown mariadb test table")
		testDb.RawQuery("DROP TABLE IF EXISTS `todos_test`;")
	}
}

type mariadbTests struct {
	name  string
	input map[string]interface{}
	want1 string
	want2 map[string]interface{}
}

func TestMariadbConstructSelectStmtFromQuerystring(t *testing.T) {
	dktest.Run(t, "mariadb:lts", mariadbOps, func(t *testing.T, c dktest.ContainerInfo) {
		ip, port, err := c.FirstPort()
		if err != nil {
			t.Fatal(err)
		}

		cfg.LoadEnvVariables()
		cfg.Vpr.Set("database.engine", "mariadb")
		cfg.Vpr.Set("database.mariadb.host", ip)
		cfg.Vpr.Set("database.mariadb.port", port)
		cfg.Vpr.Set("database.mariadb.user", "root")
		cfg.Vpr.Set("database.mariadb.pass", "root")
		cfg.Vpr.Set("database.mariadb.database", "public")
		if err := cfg.Vpr.Unmarshal(cfg); err != nil {
			log.Printf("failed loading conf, err: %+v\n", err.Error())
		}
		zlog.NewZlog()

		teardownTest := setupMariadbTestTable(t)
		defer teardownTest(t)

		var tableName = "todos_test"
		var testDb = GetDatabase(tableName, nil)
		testDb.Connect()
		tests := []mariadbTests{
			{
				name:  "get by ID",
				input: map[string]interface{}{"id": 2, "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
				want1: `SELECT * FROM todos_test WHERE id=:id ORDER BY "id" desc LIMIT 1 OFFSET 0`,
				want2: map[string]interface{}{"id": 2},
			},
			{
				name:  "get by IDs",
				input: map[string]interface{}{"id": []string{"2", "3"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
				want1: `SELECT * FROM todos_test WHERE id IN (:id1,:id2) ORDER BY "id" desc LIMIT 2 OFFSET 0`,
				want2: map[string]interface{}{"id1": "2", "id2": "3"},
			},
			{
				name:  "get keyword by ILIKE",
				input: map[string]interface{}{"task": "show", "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
				want1: `SELECT * FROM todos_test WHERE task like :task ORDER BY "id" desc LIMIT 1 OFFSET 0`,
				want2: map[string]interface{}{"task": "%show%"},
			},
			{
				name:  "get keywords by ~~ ANY(xx)",
				input: map[string]interface{}{"task": []string{"show", "stop"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}, "page": "1", "items": "5"},
				want1: `SELECT * FROM todos_test WHERE (lower(task) like :task1 or lower(task) like :task2)  ORDER BY "id" desc LIMIT 5 OFFSET 0`,
				want2: map[string]interface{}{"task1": "%show%", "task2": "%stop%"},
			},
			{
				name:  "get records by keyword that matches in given ids",
				input: map[string]interface{}{"task": "wan", "id": []string{"13", "15"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}, "page": "1", "items": "5"},
				want1: `SELECT * FROM todos_test WHERE id IN (:id1,:id2) AND task like :task ORDER BY "id" desc LIMIT 5 OFFSET 0`,
				want2: map[string]interface{}{"task": "%wan%", "id1": "13", "id2": "15"},
			},
			{
				name:  "get records by date range",
				input: map[string]interface{}{"withDateFilter": true, "created_at": "2023-01-01.2023-12-31", "columns": []string{"id", "task", "done", "created_at", "updated_at"}, "page": "1", "items": "5"},
				want1: `SELECT * FROM todos_test WHERE created_at >= :created_atFrom AND created_at <= :created_atTo ORDER BY "id" desc LIMIT 5 OFFSET 0`,
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
	})
}
