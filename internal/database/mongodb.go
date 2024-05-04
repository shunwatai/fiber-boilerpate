package database

import (
	"context"
	"fmt"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongodb struct {
	*ConnectionInfo
	TableName string
	ViewName  *string
	Db        *mongo.Client
	ctx       *context.Context
	mu        sync.Mutex
}

type MongoRows struct {
	cur *mongo.Cursor
	ctx context.Context
}

func (mr *MongoRows) StructScan(result interface{}) error {
	if err := mr.cur.Decode(result); err != nil {
		logger.Errorf("mongo decode err: %+v", err.Error())
		return err
	}
	return nil
}
func (mr *MongoRows) Next() bool {
	return mr.cur.Next(mr.ctx)
}
func (mr *MongoRows) Close() error {
	return mr.cur.Close(mr.ctx)
}

func (m *Mongodb) GetDbConfig() *ConnectionInfo {
	info, _ := GetDbConnection()
	return info
}

func (m *Mongodb) GetConnectionString() string {
	connectionString := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin&sslmode=disable", *m.User, *m.Pass, *m.Host, *m.Port, *m.Database)
	// logger.Debugf("ConnString: %+v", connectionString)
	return connectionString
}

func (m *Mongodb) Connect() {
	logger.Debugf("connecting to Mongodb... ")
	// logger.Debugf("Table: %+v", m.TableName)
	connectionString := m.GetConnectionString()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		logger.Errorf("failed to conenct mongo: %+v", err)
	}
	m.Db = client
}

// Useless for mongodb, but needed for implement the IDatabase...
func (m *Mongodb) constructSelectStmtFromQuerystring(
	queries map[string]interface{},
) (string, *helper.Pagination, map[string]interface{}) {
	return "", &helper.Pagination{}, map[string]interface{}{}
}

// return select statment and *pagination by the req querystring
func (m *Mongodb) getConditionsFromQuerystring(
	queries map[string]interface{},
	countFunc func(interface{}) (int64, error),
	// ) (string, *helper.Pagination, map[string]interface{}) {
) (bson.D, *options.FindOptions, *helper.Pagination) {
	if queries["columns"] == nil {
		logger.Errorf("queries[\"columns\"] cannot be nil...")
	}

	exactMatchCols := map[string]bool{"id": true, "_id": true} // default id(PK) & _id(mongo) have to be exact match
	// logger.Debugf("mongo query: %+v",queries)
	if queries["exactMatch"] != nil {
		for k := range queries["exactMatch"].(map[string]bool) {
			exactMatchCols[k] = true
		}
	}

	// bindvarMap := map[string]interface{}{}
	if queries["columns"] == nil {
		logger.Errorf("error: queries[\"columns\"] is nil")
	}
	cols := queries["columns"]
	pagination := helper.GetPagination(queries)
	dateRangeStmt := getDateRangeBson(queries)
	// logger.Debugf("dateRangeStmt: %+v, len: %+v", dateRangeStmt, len(dateRangeStmt))
	helper.SanitiseQuerystring(cols.([]string), queries)

	selectStmt := bson.D{}

	logger.Debugf("queries: %+v, len: %+v", queries, len(queries))
	if len(queries) != 0 || len(dateRangeStmt) != 0 { // add where clause
		// whereClauses := []string{}
		whereClauses := bson.D{}
		for k, v := range queries {
			logger.Debugf("%+v: %+v(%T)", k, v, v)
			switch v.(type) {
			case []string:
				// placeholders := []string{}
				if exactMatchCols[k] {
					oids := []primitive.ObjectID{}
					if strings.Contains(k, "_id") {
						for _, value := range v.([]string) {
							oid, _ := primitive.ObjectIDFromHex(value)
							oids = append(oids, oid)
						}
						whereClauses = append(whereClauses, bson.E{k, bson.D{{"$in", oids}}})
					} else {
						whereClauses = append(whereClauses, bson.E{k, bson.D{{"$in", v.([]string)}}})
					}
					break
				}

				// mongo $or ref: https://stackoverflow.com/a/58359960
				orWildcard := bson.A{}
				for _, value := range v.([]string) {
					orWildcard = append(orWildcard, bson.D{{k, primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", value), Options: "i"}}})
				}
				whereClauses = append(whereClauses, bson.E{"$or", orWildcard})
			default:
				if exactMatchCols[k] {
					if k == "id" {
						oid, _ := primitive.ObjectIDFromHex(v.(string))
						whereClauses = append(whereClauses, bson.E{"_id", oid})
					} else if strings.Contains(k, "_id") {
						oid, _ := primitive.ObjectIDFromHex(v.(string))
						whereClauses = append(whereClauses, bson.E{k, oid})
					} else {
						whereClauses = append(whereClauses, bson.E{k, v})
					}
					break
				}

				whereClauses = append(whereClauses, bson.E{k, primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", v), Options: "i"}})
			}
		}

		if len(dateRangeStmt) > 0 {
			whereClauses = append(whereClauses, dateRangeStmt...)
		}
		selectStmt = append(selectStmt, whereClauses...)
	}

	if count, err := countFunc(selectStmt); err != nil {
		logger.Errorf("count error: %+v", err)
	} else {
		logger.Debugf("count: %+v", count)
		pagination.Count = count
		if pagination.Items > 0 && pagination.Count > 0 {
			pagination.TotalPages = int64(math.Ceil(float64(pagination.Count) / float64(pagination.Items)))
		}
		logger.Debugf("pagination: %+v", pagination)
	}

	var limit int64
	var offset int64 = (pagination.Page - 1) * pagination.Items
	if pagination.Items == 0 {
		pagination.Items = pagination.Count
		limit = pagination.Count
	} else {
		limit = pagination.Items
	}
	options := options.Find()
	options.SetSkip(offset)
	options.SetLimit(limit)
	options.SetSort(
		bson.M{
			pagination.OrderBy["key"]: func() int {
				if pagination.OrderBy["by"] == "desc" {
					return -1
				}
				return 1
			}(),
		},
	)

	pagination.SetPageUrls()

	return selectStmt, options, pagination
}

func (m *Mongodb) Select(queries map[string]interface{}) (Rows, *helper.Pagination) {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.Debugf("select from Mongodb, table: %+v", m.TableName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.Connect()
	defer m.Db.Disconnect(ctx)

	var tableName string
	if m.ViewName != nil {
		tableName = *m.ViewName
	} else {
		tableName = m.TableName
	}

	collection := m.Db.Database(fmt.Sprintf("%s", *m.Database)).Collection(fmt.Sprintf("%s", tableName))

	var (
		cur *mongo.Cursor
		err error
	)
	// logger.Debugf("len(queries):%+v, %+v", queries, len(queries))

	var countFunc = func(filter interface{}) (int64, error) {
		count, err := collection.CountDocuments(ctx, filter)
		return count, err
	}

	conditions, findOptions, pagination := m.getConditionsFromQuerystring(queries, countFunc)
	logger.Debugf("m conditions: %+v", conditions)
	cur, err = collection.Find(ctx, conditions, findOptions)

	if err != nil {
		logger.Errorf("failed to Find, err: %+v", err)
	}

	return &MongoRows{cur, ctx}, pagination
}

func (m *Mongodb) Save(records Records) (Rows, error) {
	logger.Debugf("save from Mongodb, table: %+v", m.TableName)
	// logger.Debugf("records: %+v", records)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.Connect()
	defer m.Db.Disconnect(ctx)
	collection := m.Db.Database(fmt.Sprintf("%s", *m.Database)).Collection(fmt.Sprintf("%s", m.TableName))

	opts := options.Update().SetUpsert(true)

	upsertedIds := []string{}
	recordsMap := records.StructToMap()
	for _, record := range recordsMap {
		filter := bson.D{}
		logger.Debugf("record: %+v", record)
		if record["_id"] != nil {
			id, _ := primitive.ObjectIDFromHex(record["_id"].(string))
			filter = bson.D{{Key: "_id", Value: id}}
			record["updated_at"] = time.Now()
			if record["created_at"] != nil {
				createdAt, _ := time.Parse(time.RFC3339, record["created_at"].(string))
				record["created_at"] = createdAt
			} else {
				delete(record, "created_at")
			}
			delete(record, "_id")
			upsertedIds = append(upsertedIds, id.Hex())
		} else {
			filter = bson.D{{Key: "_id", Value: primitive.NewObjectID()}}
			record["updated_at"] = time.Now()
			record["created_at"] = time.Now()
		}

		// reserve createdBy userId
		if record["user_id"] == nil {
			delete(record, "user_id")
		}

		res, err := collection.UpdateOne(ctx, filter, bson.D{
			{Key: "$set", Value: record},
		}, opts)
		if err != nil {
			return nil, logger.Errorf("save error: %+v", err)
		}

		/* only new created records has res.UpsertedID, existing's Ids appended in the if condition above */
		if res.UpsertedID == nil {
			continue
		}
		upsertedIds = append(upsertedIds, res.UpsertedID.(primitive.ObjectID).Hex())
	}
	logger.Debugf("upsertedIds: %+v", upsertedIds)

	rows, _ := m.Select(map[string]interface{}{
		"_id":     upsertedIds,
		"columns": records.GetTags("bson"),
	})
	return rows, nil
}

func (m *Mongodb) Delete(ids []string) error {
	logger.Debugf("delete ids: %+v from Mongodb, table: %+v", ids, m.TableName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.Connect()
	defer m.Db.Disconnect(ctx)
	collection := m.Db.Database(fmt.Sprintf("%s", *m.Database)).Collection(fmt.Sprintf("%s", m.TableName))
	objectIds := []primitive.ObjectID{}
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err != nil {
			logger.Errorf("ObjectIDFromHex err: %+v", err)
		} else {
			objectIds = append(objectIds, oid)
		}
	}
	filter := bson.D{{"_id", bson.D{{"$in", objectIds}}}}
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		logger.Errorf("DeleteMany err: %+v", err)
	}
	logger.Debugf("result: %+v", result)

	return nil
}

// useless for mongo, it implemented by sqlite, postgres, mariadb
func (m *Mongodb) RawQuery(sql string) *sqlx.Rows {
	return &sqlx.Rows{}
}

// mongo version of RawQuery
func (m *Mongodb) runCommands(cmds []bson.D) error {
	for _, cmd := range cmds {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.Connect()
		defer m.Db.Disconnect(ctx)
		db := m.Db.Database(fmt.Sprintf("%s", *m.Database)).Collection(fmt.Sprintf("%s", m.TableName))

		err := db.Database().RunCommand(ctx, cmd).Err()
		if err != nil {
			logger.Errorf("mongo cmd failed: %+v", err)
			return err
		}
	}
	return nil
}
