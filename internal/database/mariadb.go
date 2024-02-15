package database

import (
	"fmt"
	"golang-api-starter/internal/helper"
	"log"
	"math"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MariaDb struct {
	*ConnectionInfo
	TableName string
	db        *sqlx.DB
}

func (m *MariaDb) GetDbConfig() *ConnectionInfo {
	info, _ := GetDbConnection()
	return info
}

func (m *MariaDb) GetConnectionString() string {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", *m.User, *m.Pass, *m.Host, *m.Port, *m.Database)
	// fmt.Printf("ConnString: %+v\n", connectionString)
	return connectionString
}

func (m *MariaDb) Connect() *sqlx.DB {
	fmt.Printf("connecting to MariaDb... \n")
	// fmt.Printf("Table: %+v\n", m.TableName)
	connectionString := m.GetConnectionString()

	db, err := sqlx.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	return db
}

// return select statment and *pagination by the req querystring
func (m *MariaDb) constructSelectStmtFromQuerystring(
	queries map[string]interface{},
) (string, *helper.Pagination, map[string]interface{}) {
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
	fmt.Printf("dateRangeStmt: %+v, len: %+v\n", dateRangeStmt, len(dateRangeStmt))
	helper.SanitiseQuerystring(cols, queries)

	countAllStmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.TableName)
	selectStmt := fmt.Sprintf(`SELECT * FROM %s`, m.TableName)

	fmt.Printf("queries: %+v, len: %+v\n", queries, len(queries))
	if len(queries) != 0 || len(dateRangeStmt) != 0 { // add where clause
		whereClauses := []string{}
		for k, v := range queries {
			fmt.Printf("%+v: %+v(%T)\n", k, v, v)
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
		selectStmt = fmt.Sprintf("%s WHERE %s ", selectStmt, strings.Join(whereClauses, " AND "))
		countAllStmt = fmt.Sprintf("%s WHERE %s", countAllStmt, strings.Join(whereClauses, " AND "))
	}

	if totalRow, err := m.db.NamedQuery(countAllStmt, bindvarMap); err != nil {
		log.Printf("Queryx Count(*) err: %+v\n", err.Error())
	} else if totalRow.Next() {
		defer totalRow.Close()
		totalRow.Scan(&pagination.Count)
	}
	if pagination.Items > 0 {
		pagination.TotalPages = int64(math.Ceil(float64(pagination.Count) / float64(pagination.Items)))
	}
	// fmt.Printf("pagination: %+v\n", pagination)

	var limit string
	var offset string = strconv.Itoa(int((pagination.Page - 1) * pagination.Items))
	if pagination.Items == 0 {
		limit = strconv.Itoa(int(pagination.Count))
	} else {
		limit = strconv.Itoa(int(pagination.Items))
	}

	selectStmt = fmt.Sprintf(`%s 
			ORDER BY %s %s
			LIMIT %s OFFSET %s
		`,
		selectStmt,
		pagination.OrderBy["key"], pagination.OrderBy["by"],
		limit, offset,
	)

	return selectStmt, pagination, bindvarMap
}

// Get all columns from db by m.TableName
// func (m *MariaDb) GetColumns() []string {
// 	selectStmt := fmt.Sprintf("select * from %s limit 1;", m.TableName)
//
// 	if m.db == nil { // for run the test case
// 		m.db = m.Connect()
// 	}
//
// 	rows, err := m.db.Queryx(selectStmt)
// 	if err != nil {
// 		log.Printf("%+v\n", err)
// 	}
// 	defer rows.Close()
//
// 	cols, err := rows.Columns()
// 	if err != nil {
// 		log.Printf("%+v\n", err)
// 	}
//
// 	return cols
// }

func (m *MariaDb) Select(queries map[string]interface{}) (Rows, *helper.Pagination) {
	fmt.Printf("select from MariaDB, table: %+v\n", m.TableName)
	m.db = m.Connect()
	defer m.db.Close()

	selectStmt, pagination, bindvarMap := m.constructSelectStmtFromQuerystring(queries)
	fmt.Printf("bindvarMap: %+v\n", bindvarMap)
	fmt.Printf("selectStmt: %+v\n", selectStmt)

	// unlike postgres, mysql needs to use "Prepare" in order to scan the column into interface{} correctly instead of unexpected []int8
	// ref:https://github.com/go-sql-driver/mysql/issues/407#issuecomment-172583652
	stmt, err := m.db.PrepareNamed(selectStmt)
	rows, err := stmt.Queryx(bindvarMap)

	if err != nil {
		log.Printf("Queryx err: %+v\n", err.Error())
	}

	if rows.Err() != nil {
		log.Printf("rows.Err(): %+v\n", err.Error())
	}

	return rows, pagination
}

func (m *MariaDb) Save(records Records) (Rows, error) {
	fmt.Printf("save from MariaDB, table: %+v\n", m.TableName)
	m.db = m.Connect()
	defer m.db.Close()

	cols := records.GetTags("db")

	// fmt.Printf("cols: %+v\n", cols)
	var colWithColon, colUpdateSet []string
	for _, col := range cols {
		// use in SQL's VALUES()
		if strings.Contains(col, "_at") {
			colWithColon = append(colWithColon, fmt.Sprintf("IFNULL(:%s, CURRENT_TIMESTAMP)", col))
		} else {
			colWithColon = append(colWithColon, fmt.Sprintf(":%s", col))
		}

		// use in SQL's ON DUPLICATE KEY UPDATE
		if strings.Contains(col, "_at") {
			colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=IFNULL(VALUES(%s), CURRENT_TIMESTAMP)", col, col))
			continue
		}
		// colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=VALUES(%s)", col, col))
		colUpdateSet = append(colUpdateSet, fmt.Sprintf("%s=IFNULL(VALUES(%s), %s.%s)", col, col, m.TableName, col))
	}

	insertStmt := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) 
		ON DUPLICATE KEY UPDATE
    %s
		RETURNING id;`,
		m.TableName,
		fmt.Sprintf(strings.Join(cols[:], ",")),
		fmt.Sprintf(strings.Join(colWithColon[:], ",")),
		fmt.Sprintf(strings.Join(colUpdateSet[:], ",\n")),
	)
	fmt.Printf("%+v \n", insertStmt)

	insertedIds := []string{}
	sqlResult, err := m.db.NamedQuery(insertStmt, records)
	if err != nil {
		log.Printf("insert error: %+v\n", err)
		return nil, err
	}
	// fmt.Printf("sqlResult: %+v\n", sqlResult)

	for sqlResult.Next() {
		var id string
		err := sqlResult.Scan(&id)
		if err != nil {
			log.Fatalf("Scan: %v", err)
			return nil, err
		}
		insertedIds = append(insertedIds, id)
	}
	if len(insertedIds) == 0 { // mariadb no error throw when violated foreign key... so check the error here
		return nil, fmt.Errorf("insert error...")
	}

	fmt.Printf("insertedIds: %+v\n", insertedIds)
	rows, _ := m.Select(map[string]interface{}{
		"id":      insertedIds,
		"columns": cols,
	})

	return rows.(*sqlx.Rows), nil
}

func (m *MariaDb) Delete(ids []string) error {
	fmt.Printf("delete from MariaDB, table: %+v\n", m.TableName)
	m.db = m.Connect()
	defer m.db.Close()

	deleteStmt, args, err := sqlx.In(
		fmt.Sprintf("DELETE FROM %s WHERE id IN (?);", m.TableName),
		ids,
	)
	if err != nil {
		log.Printf("sqlx.In err: %+v\n", err.Error())
		return err
	}
	deleteStmt = m.db.Rebind(deleteStmt)
	fmt.Printf("stmt: %+v, args: %+v\n", deleteStmt, args)

	_, err = m.db.Exec(deleteStmt, args...)
	if err != nil {
		log.Printf("Delete Query err: %+v\n", err.Error())
		return err
	}

	return nil
}

func (m *MariaDb) RawQuery(sql string) *sqlx.Rows {
	fmt.Printf("raw query from Postgres\n")
	m.db = m.Connect()
	defer m.db.Close()

	rows, err := m.db.Queryx(sql)
	if err != nil {
		log.Printf("Queryx err: %+v\n", err.Error())
	}
	if rows.Err() != nil {
		log.Printf("rows.Err(): %+v\n", err.Error())
	}

	return rows
}
