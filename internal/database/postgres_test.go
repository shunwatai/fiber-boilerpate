//go:build integration

package database

import (
	"context"
	"database/sql"
	"fmt"
	"golang-api-starter/internal/helper"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/dhui/dktest"
	_ "github.com/lib/pq"
)

var (
	opts = dktest.Options{
		PortRequired: true,
		ReadyFunc:    pgReady,
		Env:          map[string]string{"POSTGRES_PASSWORD": "password"},
	}

	pgReady = func(ctx context.Context, c dktest.ContainerInfo) bool {
		ip, port, err := c.FirstPort()
		if err != nil {
			return false
		}
		connStr := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=postgres sslmode=disable", ip, port)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return false
		}

		defer db.Close()
		return db.PingContext(ctx) == nil
	}
)

func setupPgTestTable(t *testing.T) func(t *testing.T) {
	t.Logf("setup postgres test table\n")
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "postgres")
	t.Logf("conf??? %+v, %+v,%+v,%+v,%+v,\n",
		*cfg.DbConf.PostgresConf.Host,
		*cfg.DbConf.PostgresConf.Port,
		*cfg.DbConf.PostgresConf.User,
		*cfg.DbConf.PostgresConf.Pass,
		*cfg.DbConf.PostgresConf.Database,
	)
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	zlog.NewZlog()
	var testDb = GetDatabase("", nil)

	// create test table
	testDb.RawQuery(`CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
	    NEW.updated_at = now();
	    RETURN NEW;
	END;
	$$ language 'plpgsql';`)
	testDb.RawQuery(`DROP TABLE IF EXISTS todos_test;`)
	testDb.RawQuery(`DROP SEQUENCE IF EXISTS todos_test_id_seq;`)
	testDb.RawQuery(`CREATE SEQUENCE todos_test_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;`)
	testDb.RawQuery(`CREATE TABLE "public"."todos_test" (
    "id" integer DEFAULT nextval('todos_test_id_seq') NOT NULL,
    "task" character varying(255) NOT NULL,
    "done" boolean DEFAULT false NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "todos_test_pkey" PRIMARY KEY ("id")
  ) WITH (oids = false);`)
	testDb.RawQuery(`CREATE TRIGGER update_todos_test_updated_at BEFORE UPDATE ON todos_test FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();`)

	// insert dummy data
	testDb.RawQuery(`INSERT INTO todos_test (id,task,done,created_at,updated_at) VALUES
		(2,'want sleep',false,'2023-11-03','2023-11-03'),
		(3,'stop code',false,'2023-11-03','2023-11-03'),
		(13,'take shower',false,'2023-11-03','2023-11-03'),
		(15,'want sleep',false,'2023-11-03','2023-11-03'),
		(4,'want sleep',false,'2023-11-03','2023-11-03'),
		(44,'want sleep',false,'2023-11-03','2023-11-03'),
		(41,'want sleep',false,'2023-11-03','2023-11-03')
	`)

	return func(t *testing.T) {
		t.Log("teardown postgres test table")
		testDb.RawQuery(`DROP TABLE IF EXISTS "todos_test";`)
		testDb.RawQuery(`DROP SEQUENCE IF EXISTS "todos_test_id_seq";`)
	}
}

type pgTests struct {
	name  string
	input map[string]interface{}
	want1 string
	want2 map[string]interface{}
	want3 *helper.Pagination
}

func TestPgConstructSelectStmtFromQuerystring(t *testing.T) {
	dktest.Run(t, "postgres:alpine", opts, func(t *testing.T, c dktest.ContainerInfo) {
		ip, port, err := c.FirstPort()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("setup postgres test table\n")
		cfg.LoadEnvVariables()
		cfg.Vpr.Set("database.engine", "postgres")
		cfg.Vpr.Set("database.postgres.host", ip)
		cfg.Vpr.Set("database.postgres.port", port)
		cfg.Vpr.Set("database.postgres.user", "postgres")
		cfg.Vpr.Set("database.postgres.pass", "password")
		cfg.Vpr.Set("database.postgres.database", "postgres")
		if err := cfg.Vpr.Unmarshal(cfg); err != nil {
			log.Printf("failed loading conf, err: %+v\n", err.Error())
		}
		zlog.NewZlog()

		teardownTest := setupPgTestTable(t)
		defer teardownTest(t)
		var tableName = "todos_test"
		var testDb = GetDatabase(tableName, nil)
		testDb.Connect()
		log.Printf("testDb: %+v\n", testDb)

		tests := []pgTests{
			{
				name:  "get by ID",
				input: map[string]interface{}{"id": 2, "columns": []string{"id", "task", "done", "created_at", "updated_at"}},
				want1: `SELECT * FROM todos_test WHERE id=:id ORDER BY "id" desc LIMIT 1 OFFSET 0`,
				want2: map[string]interface{}{"id": 2},
				// want3: &helper.Pagination{
				// 	Page: 1, Items: 0, Count: 1, OrderBy: map[string]string{"by": "desc", "key": "id"}, TotalPages: 1,
				// },
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
				want1: `SELECT * FROM todos_test WHERE task ILIKE :task ORDER BY "id" desc LIMIT 1 OFFSET 0`,
				want2: map[string]interface{}{"task": "%show%"},
			},
			{
				name:  "get keywords by ~~ ANY(xx)",
				input: map[string]interface{}{"task": []string{"show", "stop"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}, "page": "1", "items": "5"},
				want1: `SELECT * FROM todos_test WHERE lower(task) ~~ ANY(:task) ORDER BY "id" desc LIMIT 5 OFFSET 0`,
				want2: map[string]interface{}{"task": "{%show%,%stop%}"},
			},
			{
				name:  "get records by keyword that matches in given ids",
				input: map[string]interface{}{"task": "wan", "id": []string{"13", "15"}, "columns": []string{"id", "task", "done", "created_at", "updated_at"}, "page": "1", "items": "5"},
				want1: `SELECT * FROM todos_test WHERE id IN (:id1,:id2) AND task ILIKE :task ORDER BY "id" desc LIMIT 5 OFFSET 0`,
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
					t.Errorf("got %q want %q", strings.Fields(got1), strings.Fields(testCase.want1))
				}

				if eq := reflect.DeepEqual(testCase.want2, got2); !eq {
					t.Errorf("got %+v want %+v", got2, testCase.want2)
				}

				// skip testing want3(pagination) because of the variation of the records in DB
				// if eq := reflect.DeepEqual(testCase.want3, got3); !eq {
				// 	t.Errorf("got %+v want %+v", got3, testCase.want3)
				// }
			})
		}
	})
}
