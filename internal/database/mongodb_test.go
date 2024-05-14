//go:build integration

package database

import (
	"context"
	"fmt"
	"golang-api-starter/internal/config"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/dhui/dktest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongodbOps = dktest.Options{
		PortRequired: true,
		ReadyFunc:    mongodbReady,
		Env:          map[string]string{"MONGO_INITDB_ROOT_USERNAME": "user", "MONGO_INITDB_ROOT_PASSWORD": "password"},
	}
	mongodbReady = func(ctx context.Context, c dktest.ContainerInfo) bool {
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
)

func initTestDb() *Mongodb {
	cfg := config.Cfg
	cfg.LoadEnvVariables()
	connection := cfg.DbConf.MongodbConf
	testDb := &Mongodb{
		ConnectionInfo: &ConnectionInfo{
			Driver:   cfg.DbConf.Driver,
			Host:     connection.Host,
			Port:     connection.Port,
			User:     connection.User,
			Pass:     connection.Pass,
			Database: connection.Database,
		},
		TableName: "todos_test",
	}
	return testDb
}

func setupMongodbTestTable(t *testing.T, testRecords []map[string]interface{}) func(t *testing.T) {
	t.Logf("setup mongodb test table\n")
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "mongodb")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	zlog.NewZlog()

	testDb := initTestDb()
	// create test table
	insertCmd := []bson.D{}
	insertCmd = append(insertCmd, bson.D{
		{Key: "insert", Value: testDb.TableName},
		{Key: "documents", Value: testRecords},
	})

	// insert dummy data
	testDb.runCommands(insertCmd)

	return func(t *testing.T) {
		t.Log("teardown mongodb test table")
		dropCmd := []bson.D{}
		dropCmd = append(dropCmd, bson.D{
			{Key: "drop", Value: testDb.TableName},
		})
		testDb.runCommands(dropCmd)
	}
}

type mongodbTests struct {
	name  string
	input map[string]interface{}
	want1 bson.D
}

func TestMongodbConstructSelectStmtFromQuerystring(t *testing.T) {
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

		testRecords := []map[string]interface{}{
			{"_id": primitive.NewObjectID(), "task": "want sleep", "done": false},
			{"_id": primitive.NewObjectID(), "task": "stop code", "done": false},
			{"_id": primitive.NewObjectID(), "task": "take shower", "done": false},
			{"_id": primitive.NewObjectID(), "task": "want sleep", "done": false},
			{"_id": primitive.NewObjectID(), "task": "want sleep", "done": false},
			{"_id": primitive.NewObjectID(), "task": "want sleep", "done": false},
			{"_id": primitive.NewObjectID(), "task": "want sleep", "done": false},
		}

		teardownTest := setupMongodbTestTable(t, testRecords)
		defer teardownTest(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		testDb := initTestDb()
		testDb.Connect()
		defer testDb.Db.Disconnect(ctx)

		columns := []string{"_id", "task", "done"}

		tests := []mongodbTests{
			{
				name:  "get by ID",
				input: map[string]interface{}{"_id": testRecords[0]["_id"].(primitive.ObjectID).Hex(), "columns": columns},
				want1: bson.D{{"_id", testRecords[0]["_id"]}},
			},
			{
				name:  "get by IDs",
				input: map[string]interface{}{"_id": []string{testRecords[0]["_id"].(primitive.ObjectID).Hex(), testRecords[1]["_id"].(primitive.ObjectID).Hex()}, "columns": columns},
				want1: bson.D{{
					"_id", bson.D{{"$in", []primitive.ObjectID{
						testRecords[0]["_id"].(primitive.ObjectID),
						testRecords[1]["_id"].(primitive.ObjectID),
					}}},
				}},
			},
			{
				name:  "get keyword by ILIKE",
				input: map[string]interface{}{"task": "show", "columns": columns},
				want1: bson.D{{"task", primitive.Regex{Pattern: ".*show.*", Options: "i"}}},
			},
			{
				name:  "get keywords by ~~ ANY(xx)",
				input: map[string]interface{}{"task": []string{"show", "stop"}, "page": "1", "items": "5", "columns": columns},
				want1: bson.D{{
					"$or",
					bson.A{
						bson.D{{"task", primitive.Regex{Pattern: ".*show.*", Options: "i"}}},
						bson.D{{"task", primitive.Regex{Pattern: ".*stop.*", Options: "i"}}},
					},
				}},
			},
			{
				name:  "get records by keyword that matches in given ids",
				input: map[string]interface{}{"task": "wan", "_id": []string{testRecords[3]["_id"].(primitive.ObjectID).Hex(), testRecords[4]["_id"].(primitive.ObjectID).Hex()}, "page": "1", "items": "5", "columns": columns},
				want1: bson.D{
					{"task", primitive.Regex{Pattern: ".*wan.*", Options: "i"}},
					{
						"_id", bson.D{{"$in", []primitive.ObjectID{
							testRecords[3]["_id"].(primitive.ObjectID),
							testRecords[4]["_id"].(primitive.ObjectID),
						}}},
					},
				},
			},
			{
				name:  "get records by date range",
				input: map[string]interface{}{"withDateFilter": true, "created_at": "2023-01-01.2023-12-31", "page": "1", "items": "5", "columns": columns},
				want1: bson.D{
					{"created_at", bson.D{{"$gte", primitive.NewDateTimeFromTime(time.Date(2023, time.Month(1), 1, 0, 0, 0, 0, time.UTC))}}},
					{"created_at", bson.D{{"$lte", primitive.NewDateTimeFromTime(time.Date(2023, time.Month(12), 31, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1))}}},
				},
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				var countFunc = func(filter interface{}) (int64, error) {
					collection := testDb.Db.Database(fmt.Sprintf("%s", *testDb.Database)).Collection(fmt.Sprintf("%s", testDb.TableName))
					count, err := collection.CountDocuments(ctx, filter)
					return count, err
				}

				got1, _, _ := testDb.getConditionsFromQuerystring(testCase.input, countFunc)

				sort.Slice(got1, func(i, j int) bool {
					return got1[i].Key < got1[j].Key
				})
				sort.Slice(testCase.want1, func(i, j int) bool {
					return testCase.want1[i].Key < testCase.want1[j].Key
				})

				if eq := reflect.DeepEqual(testCase.want1, got1); !eq {
					t.Errorf("got %q \nwant %q", got1, testCase.want1)
				}
			})
		}
	})

}
