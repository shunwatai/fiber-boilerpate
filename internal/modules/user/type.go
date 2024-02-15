package user

import (
	"encoding/json"
	"fmt"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/iancoleman/strcase"
)

type UserClaims struct {
	UserId    interface{} `json:"userId"`
	Username  string      `json:"username"`
	TokenType string      `json:"tokenType"`
	jwt.StandardClaims
}

type User struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id        *int64                 `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	Name      string                 `json:"name" db:"name" bson:"name,omitempty" example:"emma" validate:"required,alphanum"`
	Password  *string                `json:"password,omitempty" db:"password" bson:"password,omitempty" example:"password"`
	FirstName *string                `json:"firstName" db:"first_name" bson:"first_name,omitempty" example:"Emma"`
	LastName  *string                `json:"lastName" db:"last_name" bson:"last_name,omitempty" example:"Watson"`
	Disabled  bool                   `json:"disabled" db:"disabled" bson:"disabled,omitempty" example:"false"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at"  bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type Users []*User

func (user *User) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *user.MongoId
	} else {
		return strconv.Itoa(int(*user.Id))
	}
}

func (users Users) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, user := range users {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(user)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (users Users) rowsToStruct(rows database.Rows) []*User {
	defer rows.Close()

	records := make([]*User, 0)
	for rows.Next() {
		var user User
		err := rows.StructScan(&user)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &user)
	}

	return records
}

func (users Users) GetTags(key string) []string {
	if len(users) == 0 {
		return []string{}
	}

	return users[0].getTags(key)
}

func (users *Users) printValue() {
	for _, v := range *users {
		if v.Id != nil {
			fmt.Printf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			fmt.Printf("new --> v: %+v\n", *v)
		}
	}
}

func (user User) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(user)
	for i := 0; i < val.Type().NumField(); i++ {
		t := val.Type().Field(i)
		fieldName := t.Name

		switch jsonTag := t.Tag.Get(tag); jsonTag {
		case "-":
		case "":
			// fmt.Println(fieldName)
		default:
			parts := strings.Split(jsonTag, ",")
			name := parts[0]
			if name == "" {
				name = fieldName
			}
			// fmt.Println(name)
			cols = append(cols, name)
		}
	}
	return cols
}
