package dbmigrate

import (
	"database/sql"
	"fmt"
	"golang-api-starter/internal/config"
	db "golang-api-starter/internal/database"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var cfg = config.Cfg

func runMigration(action string, m *migrate.Migrate) error {
	if action == "migrate-up" {
		return m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	} else if action == "migrate-down" {
		return m.Steps(-1)
	} else if action == "down-to-zero" {
		return m.Down() // for running the test case
	}
	return fmt.Errorf("failed to run migrations....")
}

func DbMigrate(action, dbDriver string) error {
	var (
		err    error
		m      *migrate.Migrate
		driver database.Driver
	)
	basepath := utils.RootDir(2)
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", dbDriver)
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	log.Printf("db driver: %+v\n", cfg.DbConf.Driver)
	log.Println(strings.Repeat("*", 50))

	dbConn := db.GetDatabase("", nil)
	dbConf := dbConn.GetDbConfig()
	connectionString := dbConn.GetConnectionString()

	if cfg.DbConf.Driver == "postgres" {
		logger.Infof("connectionString: %+v", connectionString)
		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatalf("sql.Open error: %+v\n", err)
		}

		driver, err = postgres.WithInstance(db, &postgres.Config{DatabaseName: *dbConf.Database})
		if err != nil {
			log.Fatalf("postgres.WithInstance error: %+v\n", err)
		}

	} else if cfg.DbConf.Driver == "mariadb" {
		db, err := sql.Open("mysql", connectionString)
		if err != nil {
			log.Fatalf("sql.Open error: %+v\n", err)
		}

		driver, err = mysql.WithInstance(db, &mysql.Config{DatabaseName: *dbConf.Database})
		if err != nil {
			log.Fatalf("mysql.WithInstance error: %+v\n", err)
		}

	} else if cfg.DbConf.Driver == "sqlite" {
		db, err := sql.Open("sqlite3", connectionString)
		if err != nil {
			log.Fatalf("sql.Open error: %+v\n", err)
		}

		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{DatabaseName: *dbConf.Database})
		if err != nil {
			log.Fatalf("sqlite.WithInstance error: %+v\n", err)
		}

	} else if cfg.DbConf.Driver == "mongodb" {
		mongo := db.Mongodb{ConnectionInfo: dbConf}
		mongo.Connect()

		if err != nil {
			log.Fatalf("sql.Open error: %+v\n", err)
		}

		driver, err = mongodb.WithInstance(mongo.Db, &mongodb.Config{DatabaseName: *dbConf.Database})
		if err != nil {
			log.Fatalf("mongodb.WithInstance error: %+v\n", err)
		}
	}

	m, err = migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file:///%s/migrations/%s", basepath, dbDriver),
		*dbConf.Database, driver)
	if err != nil {
		log.Fatalf("migrate.NewWithDatabaseInstance error: %+v\n", err)
	}

	if err = runMigration(action, m); err != nil {
		logger.Fatalf("runMigration err: %+v", err)
		return err
	}
	ver, dir, err := m.Version()
	log.Println(strings.Repeat("*", 50))
	log.Printf("migrated success, version: %+v, \n", ver)
	log.Printf("migrated failed, dirty: %+v, \n", dir)
	log.Printf("migrated error, error: %+v, \n", err)
	log.Println(strings.Repeat("*", 50))

	return err
}
