//go:build integration

package dbmigrate

import (
	"context"
	"database/sql"
	"fmt"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhui/dktest"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type migrateTests struct {
	name  string
	input map[string]string
	want  error
}

var (
	pgOpts = dktest.Options{
		PortRequired: true,
		ReadyFunc:    pgReady,
		Env:          map[string]string{"POSTGRES_PASSWORD": "password"},
	}
	mariadbOps = dktest.Options{
		PortRequired: true,
		ReadyFunc:    mariadbReady,
		Env:          map[string]string{"MYSQL_ROOT_PASSWORD": "root", "MYSQL_DATABASE": "public"},
	}
	mongodbOps = dktest.Options{
		PortRequired: true,
		ReadyFunc:    mongodbReady,
		Env:          map[string]string{"MONGO_INITDB_ROOT_USERNAME": "user", "MONGO_INITDB_ROOT_PASSWORD": "password"},
	}
)

func pgReady(ctx context.Context, c dktest.ContainerInfo) bool {
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

func mariadbReady(ctx context.Context, c dktest.ContainerInfo) bool {
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

func mongodbReady(ctx context.Context, c dktest.ContainerInfo) bool {
	ip, port, err := c.FirstPort()
	if err != nil {
		return false
	}

	connStr := fmt.Sprintf("mongodb://user:password@%s:%s/test?authSource=admin&sslmode=disable", ip, port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	return err == nil
}

func TestDbMigrate(t *testing.T) {
	// test postgres migration up & doen
	dktest.Run(t, "postgres:alpine", pgOpts, func(t *testing.T, c dktest.ContainerInfo) {
		ip, port, err := c.FirstPort()
		if err != nil {
			t.Fatal(err)
		}

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

		tests := []migrateTests{
			{
				name:  "Test postgres migrate up",
				input: map[string]string{"action": "migrate-up", "dbDriver": "postgres"},
				want:  nil,
			},
			{
				name:  "Test postgres migrate down",
				input: map[string]string{"action": "down-to-zero", "dbDriver": "postgres"},
				want:  fmt.Errorf("no migration"),
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				got := DbMigrate(testCase.input["action"], testCase.input["dbDriver"])

				if testCase.input["action"] == "migrate-up" {
					assertNoError(t, got)
				}
				if testCase.input["action"] == "down-to-zero" {
					assertError(t, got, testCase.want)
				}
			})
		}
	})

	// test mariadb migration up & doen
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

		tests := []migrateTests{
			{
				name:  "Test mariadb migrate up",
				input: map[string]string{"action": "migrate-up", "dbDriver": "mariadb"},
				want:  nil,
			},
			{
				name:  "Test mariadb migrate down",
				input: map[string]string{"action": "down-to-zero", "dbDriver": "mariadb"},
				want:  fmt.Errorf("no migration"),
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				got := DbMigrate(testCase.input["action"], testCase.input["dbDriver"])

				if testCase.input["action"] == "migrate-up" {
					assertNoError(t, got)
				}
				if testCase.input["action"] == "down-to-zero" {
					assertError(t, got, testCase.want)
				}
			})
		}
	})

	// test mongodb migration up & doen
	dktest.Run(t, "mongo:5.0.12", mongodbOps, func(t *testing.T, c dktest.ContainerInfo) {
		ip, port, err := c.FirstPort()
		if err != nil {
			t.Fatal(err)
		}

		cfg.LoadEnvVariables()
		cfg.Vpr.Set("database.engine", "mongodb")
		cfg.Vpr.Set("database.mongodb.host", ip)
		cfg.Vpr.Set("database.mongodb.port", port)
		cfg.Vpr.Set("database.mongodb.user", "user")
		cfg.Vpr.Set("database.mongodb.pass", "password")
		cfg.Vpr.Set("database.mongodb.database", "test")
		if err := cfg.Vpr.Unmarshal(cfg); err != nil {
			log.Printf("failed loading conf, err: %+v\n", err.Error())
		}
		zlog.NewZlog()

		tests := []migrateTests{
			{
				name:  "Test mongodb migrate up",
				input: map[string]string{"action": "migrate-up", "dbDriver": "mongodb"},
				want:  nil,
			},
			{
				name:  "Test mongodb migrate down",
				input: map[string]string{"action": "down-to-zero", "dbDriver": "mongodb"},
				want:  fmt.Errorf("no migration"),
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				got := DbMigrate(testCase.input["action"], testCase.input["dbDriver"])

				if testCase.input["action"] == "migrate-up" {
					assertNoError(t, got)
				}
				if testCase.input["action"] == "down-to-zero" {
					assertError(t, got, testCase.want)
				}
			})
		}
	})

	// test sqlite migration up & doen
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
	zlog.NewZlog()

	tests := []migrateTests{
		{
			name:  "Test sqlite migrate up",
			input: map[string]string{"action": "migrate-up", "dbDriver": "sqlite"},
			want:  nil,
		},
		{
			name:  "Test sqlite migrate down",
			input: map[string]string{"action": "down-to-zero", "dbDriver": "sqlite"},
			want:  fmt.Errorf("no migration"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := DbMigrate(testCase.input["action"], testCase.input["dbDriver"])

			if testCase.input["action"] == "migrate-up" {
				assertNoError(t, got)
			}
			if testCase.input["action"] == "down-to-zero" {
				assertError(t, got, testCase.want)
			}
		})
	}
}

func assertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Fatal("got an error but didn't want one")
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if got == nil {
		t.Fatal("didn't get an error but wanted one")
	}

	if got.Error() != want.Error() {
		t.Errorf("got %q, want %q", got, want)
	}
}
