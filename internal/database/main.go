package database

import (
	"fmt"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rows interface {
	StructScan(interface{}) error
	Next() bool
	Close() error
}

type IDatabase interface {
	/* Initiate the DB connection to its correspond struct */
	Connect()

	/* Get ConnectionString */
	GetDbConfig() *ConnectionInfo

	/* Get ConnectionString */
	GetConnectionString() string

	/* Select by raw sql */
	RawQuery(string) *sqlx.Rows

	/* Select by req querystring with pagination */
	Select(map[string]interface{}) (Rows, *helper.Pagination)

	// Insert new records, support upsert when id is present.
	// And also support batch insert/upsert
	Save(Records) (Rows, error)

	/* Delete records by ids(support batch delete) */
	Delete([]string) error

	constructSelectStmtFromQuerystring(queries map[string]interface{}) (string, *helper.Pagination, map[string]interface{})
}

type ConnectionInfo struct {
	Driver   string
	Host     *string
	Port     *string
	User     *string
	Pass     *string
	Database *string
}

type Records interface {
	StructToMap() []map[string]interface{}
	GetTags(string) []string
}

type IgnoredCols []string
var IgnrCols = new(IgnoredCols)
func SetIgnoredCols(cols ...string) {
	*IgnrCols = IgnoredCols(cols)
}

var cfg = config.Cfg

func GetDbConnection() (*ConnectionInfo, error) {
	if cfg.DbConf.Driver == "sqlite" {
		connection := cfg.DbConf.SqliteConf
		return &ConnectionInfo{
			Driver:   cfg.DbConf.Driver,
			Host:     connection.Host,
			Port:     connection.Port,
			User:     connection.User,
			Pass:     connection.Pass,
			Database: connection.Database,
		}, nil
	}
	if cfg.DbConf.Driver == "postgres" {
		connection := cfg.DbConf.PostgresConf
		return &ConnectionInfo{
			Driver:   cfg.DbConf.Driver,
			Host:     connection.Host,
			Port:     connection.Port,
			User:     connection.User,
			Pass:     connection.Pass,
			Database: connection.Database,
		}, nil
	}
	if cfg.DbConf.Driver == "mariadb" {
		connection := cfg.DbConf.MariadbConf
		return &ConnectionInfo{
			Driver:   cfg.DbConf.Driver,
			Host:     connection.Host,
			Port:     connection.Port,
			User:     connection.User,
			Pass:     connection.Pass,
			Database: connection.Database,
		}, nil
	}
	if cfg.DbConf.Driver == "mongodb" {
		connection := cfg.DbConf.MongodbConf
		return &ConnectionInfo{
			Driver:   cfg.DbConf.Driver,
			Host:     connection.Host,
			Port:     connection.Port,
			User:     connection.User,
			Pass:     connection.Pass,
			Database: connection.Database,
		}, nil
	}
	return nil, fmt.Errorf("GetDbConnection() error, please check the cfg.DbConf.Driver")
}

func GetDatabase(tableName string, viewName *string) IDatabase {
	if cfg.DbConf == nil {
		logger.Errorf("error: DbConf is nil, maybe fail to load the config....")
	}
	logger.Debugf("engine: %+v", cfg.DbConf.Driver)

	if cfg.DbConf.Driver == "sqlite" {
		dbConn, _ := GetDbConnection()
		return &Sqlite{
			ConnectionInfo: dbConn,
			TableName:      tableName,
			ViewName:       viewName,
		}
	}

	if cfg.DbConf.Driver == "mariadb" {
		dbConn, _ := GetDbConnection()
		return &MariaDb{
			ConnectionInfo: dbConn,
			TableName:      tableName,
			ViewName:       viewName,
		}
	}

	if cfg.DbConf.Driver == "postgres" {
		dbConn, _ := GetDbConnection()
		return &Postgres{
			ConnectionInfo: dbConn,
			TableName:      tableName,
			ViewName:       viewName,
		}
	}

	if cfg.DbConf.Driver == "mongodb" {
		dbConn, _ := GetDbConnection()
		return &Mongodb{
			ConnectionInfo: dbConn,
			TableName:      tableName,
			ViewName:       viewName,
		}
	}

	return nil
}

// Generate partial sql(bindvar) for date range filter by the conditions in parameter "queries",
// the date's key will be removed in "queries" after processed.
//
// NOTE: when getting the querystring in controller, GetQueryString() will look for keys end with either "_at" or "date" in order to add the flag "withDateFilter" in queries.
//
// There are 3 accepted input, all indicated by .(DOT):
//  1. records from specified date up to now --> 2023-10-1.
//     e.g. queries --> map[string]interface{}{"created_at":"2023-10-01."} --> sql: created_at >= 2023-10-01
//  2. records before specified date --> .2023-06-30
//     e.g. queries --> map[string]interface{}{"created_at":".2023-06-30"} --> sql: created_at <= 2023-06-30
//  3. records in between 2 dates --> 2023-06-30.2023-10-1
//     e.g. queries --> map[string]interface{}{"created_at":"2023-06-30.2023-10-01"} --> sql: created_at >= 2023-06-30 AND created_at <= 2023-10-01
func getDateRangeStmt(queries, bindvarMap map[string]interface{}) string {
	// logger.Debugf("dd query: %+v\n", queries)
	if queries["withDateFilter"] == nil {
		return ""
	}
	dateRangeConditions := []string{}
	for k, v := range queries {
		if len(k) < 3 || (!strings.Contains(k[len(k)-4:], "date") && !strings.Contains(k[len(k)-3:], "_at")) {
			// logger.Debugf("not date: %+v\n", k)
			continue
		}
		splitedDates := strings.Split(v.(string), ".")
		// logger.Debugf("splitedDates? %+v, len: %+v\n", splitedDates, len(splitedDates))
		if len(splitedDates) == 2 {
			from, to := splitedDates[0], splitedDates[1]
			if from != "" {
				fromKey := fmt.Sprintf("%sFrom", k)
				dateRangeConditions = append(dateRangeConditions, fmt.Sprintf("%s >= :%s", k, fromKey))
				bindvarMap[fromKey] = from
			}
			if to != "" {
				toKey := fmt.Sprintf("%sTo", k)
				dateRangeConditions = append(dateRangeConditions, fmt.Sprintf("%s <= :%s", k, toKey))
				bindvarMap[toKey] = to
			}
		}
		delete(queries, k)
	}

	// logger.Debugf("dateConditions: %+v\n",dateConditions)
	return strings.Join(dateRangeConditions, " AND ")
}

// Generate bson for mongo find's date filtering
func getDateRangeBson(queries map[string]interface{}) bson.D {
	// logger.Debugf("dd query: %+v\n", queries)
	if queries["withDateFilter"] == nil {
		return bson.D{}
	}

	const dateFormat = "2006-01-02"
	dateRangeConditions := bson.D{}
	for k, v := range queries {
		if len(k) < 3 || (!strings.Contains(k[len(k)-4:], "date") && !strings.Contains(k[len(k)-3:], "_at")) {
			// logger.Debugf("not date: %+v\n", k)
			continue
		}
		splitedDates := strings.Split(v.(string), ".")
		// logger.Debugf("splitedDates? %+v, len: %+v\n", splitedDates, len(splitedDates))
		if len(splitedDates) == 2 {
			from, to := splitedDates[0], splitedDates[1]
			if from != "" {
				t, _ := time.Parse(dateFormat, from)
				dateRangeConditions = append(dateRangeConditions, bson.D{{
					k, bson.D{{
						"$gte", primitive.NewDateTimeFromTime(t),
					}},
				}}...)
			}
			if to != "" {
				t, _ := time.Parse(dateFormat, to)
				dateRangeConditions = append(dateRangeConditions, bson.D{{
					k, bson.D{{
						"$lte", primitive.NewDateTimeFromTime(t.AddDate(0, 0, 1)),
					}},
				}}...)
			}
		}
		delete(queries, k)
	}

	// logger.Debugf("dateConditions: %+v\n",dateConditions)
	// return strings.Join(dateRangeConditions, " AND ")
	return dateRangeConditions
}

// GetIdsMapCondition() return map[string]interface{} sth like map[string]interface{}{"id":[]string{x,y,z}}
// which can use for Get() to retrieve records by id(s)
// param "keyId" can be nil for default "id" in db, it can be other column(foreign key) like "todo_id"
func GetIdsMapCondition(keyId *string, ids []string) map[string]interface{} {
	condition := map[string]interface{}{}

	if keyId == nil {
		if cfg.DbConf.Driver == "mongodb" {
			condition["_id"] = ids
		} else {
			condition["id"] = ids
		}
	} else {
		condition[*keyId] = ids
	}

	return condition
}
