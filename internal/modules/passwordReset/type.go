package passwordReset

import (
	"encoding/json"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"slices"

	//"golang-api-starter/internal/modules/user"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/iancoleman/strcase"
)

type ResetClaims struct {
	UserId    interface{} `json:"userId"`
	Email     string      `json:"email"`
	TokenType string      `json:"tokenType"`
	jwt.RegisteredClaims
}

type PasswordReset struct {
	MongoId *string     `json:"_id,omitempty" bson:"_id,omitempty" validate:"omitempty,id_custom_validation"` // https://stackoverflow.com/a/20739427
	Id      *int64      `json:"id" db:"id" bson:"id,omitempty" example:"2" validate:"omitempty,id_custom_validation"`
	UserId  interface{} `json:"userId" db:"user_id" bson:"user_id,omitempty" validate:"omitempty,id_custom_validation"`
	//User      *user.User             `json:"user"`
	Email      string                 `json:"email,omitempty"`
	Token      string                 `json:"token,omitempty"`
	TokenHash  *string                `json:"tokenHash,omitempty" db:"token_hash" bson:"token_hash,omitempty"`
	ExpiryDate *helper.CustomDatetime `json:"expiryDate" db:"expiry_date" bson:"expiry_date,omitempty"`
	IsUsed     bool                   `json:"isUsed" db:"is_used" bson:"is_used,omitempty" example:"false"`
	CreatedAt  *helper.CustomDatetime `json:"createdAt" db:"created_at" bson:"created_at,omitempty"`
	UpdatedAt  *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
}

type PasswordResets []*PasswordReset

func (pr *PasswordReset) GetId() string {
	if cfg.DbConf.Driver == "mongodb" {
		return *pr.MongoId
	} else {
		return strconv.Itoa(int(*pr.Id))
	}
}

func (pr *PasswordReset) GetUserId() string {
	if cfg.DbConf.Driver == "mongodb" {
		userId, ok := pr.UserId.(string)
		if !ok {
			return ""
		}
		return userId
	} else {
		return strconv.Itoa(int(pr.UserId.(int64)))
	}
}

func (prs PasswordResets) StructToMap() []map[string]interface{} {
	mapsResults := []map[string]interface{}{}
	for _, pr := range prs {
		tmp := map[string]interface{}{}
		result := map[string]interface{}{}
		data, _ := json.Marshal(pr)
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			result[strcase.ToSnake(k)] = v
		}
		mapsResults = append(mapsResults, result)
	}

	return mapsResults
}

func (prs PasswordResets) rowsToStruct(rows database.Rows) []*PasswordReset {
	defer rows.Close()

	records := make([]*PasswordReset, 0)
	for rows.Next() {
		var pr PasswordReset
		err := rows.StructScan(&pr)
		if err != nil {
			log.Fatalf("Scan: %v", err)
		}
		records = append(records, &pr)
	}

	return records
}

func (prs PasswordResets) GetTags(key string) []string {
	if len(prs) == 0 {
		return []string{}
	}

	return prs[0].getTags(key)
}

func (prs *PasswordResets) printValue() {
	for _, v := range *prs {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
		}
	}
}

// get the tags by key(json / db / bson) name from the struct
// ref: https://stackoverflow.com/a/40865028
func (pr PasswordReset) getTags(key ...string) []string {
	var tag string
	if len(key) == 1 {
		tag = key[0]
	} else if cfg.DbConf.Driver == "mongodb" {
		tag = "bson"
	} else {
		tag = "db"
	}

	cols := []string{}
	val := reflect.ValueOf(pr)
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
			if !slices.Contains(*database.IgnrCols, name) {
				cols = append(cols, name)
			}
		}
	}
	return cols
}
