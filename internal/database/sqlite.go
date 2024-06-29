package database

import (
	"flag"
	"fmt"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	*ConnectionInfo
	TableName string
	ViewName  *string
	db        *sqlx.DB
	mu        sync.Mutex
}

// Get all columns from db by m.TableName
// func (m *Sqlite) GetColumns() []string {
// 	selectStmt := fmt.Sprintf("select * from %s limit 1;", m.TableName)
//
// 	if m.db == nil { // for run the test case
// 		m.db = m.Connect()
// 	}
//
// 	rows, err := m.db.Queryx(selectStmt)
// 	if err != nil {
// 		logger.Errorf("Queryx err: %+v", err)
// 	}
// 	defer rows.Close()
//
// 	cols, err := rows.Columns()
// 	if err != nil {
// 		logger.Errorf("Failed to get columns err: %+v", err)
// 	}
//
// 	return cols
// }

func (m *Sqlite) GetDbConfig() *ConnectionInfo {
	info, _ := GetDbConnection()
	return info
}

func (m *Sqlite) GetConnectionString() string {
	var dbFile, connectionString string
	// sqlite db get wrong path when running test, so need to ../../
	// ref: https://stackoverflow.com/a/36666114
	if flag.Lookup("test.v") == nil {
		logger.Infof("normal run")
		dbFile = fmt.Sprintf("%s.db", *m.Database)
		connectionString = fmt.Sprintf("./%s?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha1&parseTime=true", dbFile, *m.User, *m.Pass)
	} else {
		logger.Infof("run under go test")
		connectionString = *m.Database
	}
	// logger.Debugf("ConnString: %+v", connectionString)
	// os.Remove(dbFile)

	return connectionString
}

func (m *Sqlite) Connect() {
	logger.Debugf("connecting to Sqlite... ")
	logger.Debugf("Table: %+v", m.TableName)

	connectionString := m.GetConnectionString()
	db, err := sqlx.Open("sqlite3", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	m.db = db
}

// return select statment and *pagination by the req querystring
func (m *Sqlite) constructSelectStmtFromQuerystring(
	queries map[string]interface{},
) (string, *helper.Pagination, map[string]interface{}) {
	if queries["columns"] == nil {
		logger.Errorf("queries[\"columns\"] cannot be nil...")
	}

	var tableName string
	if m.ViewName != nil {
		tableName = *m.ViewName
	} else {
		tableName = m.TableName
	}

	exactMatchCols := map[string]bool{"id": true} // default id(PK) have to be exact match
	if queries["exactMatch"] != nil {
		for k := range queries["exactMatch"].(map[string]bool) {
			exactMatchCols[k] = true
		}
	}

	bindvarMap := map[string]interface{}{}
	cols := queries["columns"].([]string)

	pagination := helper.GetPagination(queries)
	dateRangeStmt := getDateRangeStmt(queries, bindvarMap)
	logger.Debugf("dateRangeStmt: %+v, len: %+v", dateRangeStmt, len(dateRangeStmt))
	helper.SanitiseQuerystring(cols, queries)

	countAllStmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	selectStmt := fmt.Sprintf(`SELECT * FROM %s`, tableName)

	logger.Debugf("queries: %+v, len: %+v", queries, len(queries))
	if len(queries) != 0 || len(dateRangeStmt) != 0 { // add where clause
		whereClauses := []string{}
		for k, v := range queries {
			logger.Debugf("%+v: %+v(%T)", k, v, v)
			switch v.(type) {
			case []string:
				placeholders := []string{}
				if exactMatchCols[k] || strings.Contains(k, "_id") {
					for i, value := range v.([]string) {
						key := fmt.Sprintf(":%s%d", k, i+1)
						bindvarMap[key[1:]] = value
						placeholders = append(placeholders, key)
					}
					whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)",
						k, strings.ToLower(strings.Join(placeholders, ",")),
					))
					break
				}

				multiLikeClause := []string{}
				for i, value := range v.([]string) {
					key := fmt.Sprintf("%s%d", k, i+1)
					bindvarMap[key] = fmt.Sprintf("%%%s%%", value)
					multiLikeClause = append(multiLikeClause, fmt.Sprintf("lower(%s) LIKE :%s", k, key))
				}
				whereClauses = append(whereClauses,
					fmt.Sprintf("(%s)", strings.ToLower(strings.Join(multiLikeClause, " OR "))),
				)
			default:
				if exactMatchCols[k] || strings.Contains(k, "_id") {
					bindvarMap[k] = v
					whereClauses = append(whereClauses, fmt.Sprintf("%s=:%s", k, k))
					break
				}

				bindvarMap[k] = fmt.Sprintf("%%%s%%", v)
				whereClauses = append(whereClauses, strings.ToLower(fmt.Sprintf("%s LIKE :%s", k, k)))
			}
		}

		if len(dateRangeStmt) > 0 {
			whereClauses = append(whereClauses, dateRangeStmt)
		}
		slices.Sort(whereClauses) // useless, just to avoid assert error in sqlite_test.go
		selectStmt = fmt.Sprintf("%s WHERE %s", selectStmt, strings.Join(whereClauses, " AND "))
		countAllStmt = fmt.Sprintf("%s WHERE %s", countAllStmt, strings.Join(whereClauses, " AND "))
	}

	if totalRow, err := m.db.NamedQuery(countAllStmt, bindvarMap); err != nil {
		logger.Debugf("Queryx Count(*) err: %+v", err.Error())
	} else if totalRow.Next() {
		defer totalRow.Close()
		totalRow.Scan(&pagination.Count)
	}
	if pagination.Items > 0 && pagination.Count > 0 {
		pagination.TotalPages = int64(math.Ceil(float64(pagination.Count) / float64(pagination.Items)))
	}
	logger.Debugf("pagination: %+v", pagination)

	var limit string
	var offset string = strconv.Itoa(int((pagination.Page - 1) * pagination.Items))
	if pagination.Items == 0 {
		pagination.Items = pagination.Count
		limit = strconv.Itoa(int(pagination.Count))
	} else {
		limit = strconv.Itoa(int(pagination.Items))
	}

	selectStmt = fmt.Sprintf(`%s 
			ORDER BY "%s" %s
			LIMIT %s OFFSET %s
		`,
		selectStmt,
		pagination.OrderBy["key"], pagination.OrderBy["by"],
		limit, offset,
	)

	pagination.SetPageUrls()

	return selectStmt, pagination, bindvarMap
}

// func (m *Sqlite) GetConnectionInfo() ConnectionInfo {
// 	return *m.ConnectionInfo
// }

func (m *Sqlite) Select(queries map[string]interface{}) (Rows, *helper.Pagination) {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.Debugf("select from Sqlite, table: %+v", m.TableName)
	m.Connect()
	defer m.db.Close()

	selectStmt, pagination, bindvarMap := m.constructSelectStmtFromQuerystring(queries)
	logger.Debugf("bindvarMap: %+v", bindvarMap)
	logger.Debugf("selectStmt: %+v", selectStmt)

	rows, err := m.db.NamedQuery(selectStmt, bindvarMap)
	if err != nil {
		logger.Errorf("Queryx err: %+v", err.Error())
	}

	if rows.Err() != nil {
		logger.Errorf("rows.Err(): %+v", err.Error())
	}

	return rows, pagination
}

func (m *Sqlite) Save(records Records) (Rows, error) {
	logger.Debugf("save from Sqlite, table: %+v", m.TableName)
	// logger.Debugf("records: %+v", records)
	m.Connect()
	defer m.db.Close()

	cols := records.GetTags("db")

	// logger.Debugf("cols: %+v", cols)
	var colWithColon, colUpdateSet []string
	for _, col := range cols {
		// use in SQL's VALUES()
		if strings.Contains(col, "_at") {
			colWithColon = append(colWithColon, fmt.Sprintf("IFNULL(:%s, CURRENT_TIMESTAMP)", col))
		} else {
			colWithColon = append(colWithColon, fmt.Sprintf(":%s", col))
		}

		// use in SQL's ON CONFLICT DO UPDATE SET
		if strings.Contains(col, "_at") {
			colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=IFNULL(excluded.%s, CURRENT_TIMESTAMP)", col, col))
			continue
		}
		// colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=IFNULL(excluded.%s, %s.%s)", col, col, m.TableName, col))
		colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=excluded.%s", col, col))
	}

	insertStmt := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) 
		ON CONFLICT(id) DO UPDATE SET
    %s
		RETURNING id`,
		m.TableName,
		fmt.Sprintf(strings.Join(cols[:], ",")),
		fmt.Sprintf(strings.Join(colWithColon[:], ",")),
		fmt.Sprintf(strings.Join(colUpdateSet[:], ",")),
	)
	logger.Debugf("%+v ", insertStmt)

	// no idea why sqlite batch insert cannot retrieve all ids, so insert one by one in loop
	insertedIds := []string{}
	mapsResults := records.StructToMap()
	for _, record := range mapsResults {
		logger.Debugf("record: %+v ", record)
		sqlResult, err := m.db.NamedExec(insertStmt, record)
		if err != nil {
			logger.Errorf("insert error: %+v", err)
			return nil, err
		}
		lastId, _ := sqlResult.LastInsertId()

		if record["id"] != nil {
			insertedIds = append(insertedIds, strconv.Itoa(int(record["id"].(float64))))
			continue
		}
		insertedIds = append(insertedIds, strconv.Itoa(int(lastId)))
	}

	logger.Debugf("insertedIds: %+v", insertedIds)
	rows, _ := m.Select(map[string]interface{}{
		"id":      insertedIds,
		"columns": cols,
	})

	return rows, nil
}

// func (m *Sqlite) Update(records Records) *sqlx.Rows {
// 	logger.Debugf("update from Sqlite, table: %+v", m.TableName)
// 	return &sqlx.Rows{}
// }

func (m *Sqlite) Delete(ids []string) error {
	logger.Debugf("delete from Sqlite, table: %+v where id in (%+v)", m.TableName, ids)
	m.Connect()
	defer m.db.Close()

	deleteStmt, args, err := sqlx.In(
		fmt.Sprintf("DELETE FROM %s WHERE id IN (?);", m.TableName),
		ids,
	)
	if err != nil {
		logger.Errorf("sqlx.In err: %+v", err.Error())
		return err
	}
	deleteStmt = m.db.Rebind(deleteStmt)
	logger.Debugf("stmt: %+v, args: %+v", deleteStmt, args)

	_, err = m.db.Exec(deleteStmt, args...)
	if err != nil {
		logger.Errorf("Delete Query err: %+v", err.Error())
		return err
	}

	return nil
}

func (m *Sqlite) RawQuery(sql string) *sqlx.Rows {
	logger.Debugf("raw query from Sqlite")
	m.Connect()
	defer m.db.Close()

	// hack: Queryx cannot run CREATE or INSERT statement for sqlite, so use Exec()
	if !strings.Contains(strings.ToLower(sql), "select") {
		m.db.Exec(sql)
	}

	rows, err := m.db.Queryx(sql)
	if err != nil {
		logger.Errorf("Queryx err: %+v", err.Error())
	}
	if rows.Err() != nil {
		logger.Errorf("rows.Err(): %+v", err.Error())
	}

	return rows
}
