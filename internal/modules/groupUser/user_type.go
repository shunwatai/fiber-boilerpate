package groupUser

import (
	"encoding/json"
	"errors"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"slices"

	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

type UserDto struct {
	MongoId   helper.Optional[string]                `json:"_id,omitempty"`
	Id        helper.Optional[helper.FlexInt]        `json:"id" example:"2"`
	Name      helper.Optional[string]                `json:"name" example:"emma"`
	Password  helper.Optional[string]                `json:"password,omitempty" example:"password"`
	Email     helper.Optional[string]                `json:"email"  example:"xxx@example.com"`
	FirstName helper.Optional[string]                `json:"firstName" example:"Emma"`
	LastName  helper.Optional[string]                `json:"lastName" example:"Watson"`
	Disabled  helper.Optional[bool]                  `json:"disabled" example:"false"`
	IsOauth   helper.Optional[bool]                  `json:"isOauth" example:"false"`
	Provider  helper.Optional[string]                `json:"provider,omitempty" example:"google"`
	Groups    helper.Optional[[]*Group]              `json:"groups"`
	CreatedAt helper.Optional[helper.CustomDatetime] `json:"createdAt"`
	UpdatedAt helper.Optional[helper.CustomDatetime] `json:"updatedAt"`
	Search    helper.Optional[string]                `json:"-" example:"google"`
}
type UsersDto []*UserDto

func (userDto *UserDto) Validate(action string) error {
	var ignoreRequiredCheckIfUpdate = func(presented bool) bool {
		if action == "update" {
			return true
		}
		return presented
	}
	validate := validator.New()
	var validateErrs []error
	var validations = map[string]map[bool]map[string]map[string]any{
		"password": {
			ignoreRequiredCheckIfUpdate(userDto.Password.Presented): {
				"omitempty,min=4": {"at least 4 character": userDto.Password.Value},
			},
		},
		"name": {
			ignoreRequiredCheckIfUpdate(userDto.Name.Presented): {
				"omitempty,alphanum": {"only allow A-Z1-9, no space": userDto.Name.Value},
			},
		},
		"email": {
			ignoreRequiredCheckIfUpdate(userDto.Email.Presented): {
				"omitempty,email": {"is not valid": userDto.Email.Value},
			},
		},
	}
	for key, presentedRuleValue := range validations {
		for presented, ruleErrmsgValue := range presentedRuleValue {
			if !presented {
				validateErrs = append(validateErrs, errors.New(key+" is required"))
			} else {
				for rule, errmsgValue := range ruleErrmsgValue {
					for errMsg, value := range errmsgValue {
						if err := validate.Var(value, rule); err != nil {
							validateErrs = append(validateErrs, errors.New(key+" "+errMsg))
						}
					}
				}
			}
		}
	}

	return errors.Join(validateErrs...)
}

func (userDto *UserDto) GetId() string {
	if cfg.DbConf.Driver == "mongodb" && userDto.MongoId.Presented {
		return *userDto.MongoId.Value
	} else if userDto.Id.Presented {
		return strconv.Itoa(int(*userDto.Id.Value))
	} else {
		return ""
	}
}

func (ud *UserDto) MapToUser(user *User) {
	if ud.MongoId.Presented {
		user.MongoId = ud.MongoId.Value
	}
	if ud.Id.Presented {
		user.Id = ud.Id.Value
	}
	if ud.Name.Presented {
		user.Name = *ud.Name.Value
	}
	if ud.Password.Presented && ud.Password.Value != nil && len(*ud.Password.Value) > 0 {
		user.Password = utils.ToPtr(utils.HashPassword(*ud.Password.Value))
	}
	if ud.Email.Presented {
		user.Email = ud.Email.Value
	}
	if ud.FirstName.Presented {
		user.FirstName = ud.FirstName.Value
	}
	if ud.LastName.Presented {
		user.LastName = ud.LastName.Value
	}
	if ud.Disabled.Presented {
		user.Disabled = *ud.Disabled.Value
	}
	if ud.IsOauth.Presented {
		user.IsOauth = *ud.IsOauth.Value
	}
	if ud.Provider.Presented {
		user.Provider = ud.Provider.Value
	}
	if ud.Groups.Presented {
		user.Groups = *ud.Groups.Value
	}
	if ud.CreatedAt.Presented {
		user.CreatedAt = ud.CreatedAt.Value
	}
	if ud.UpdatedAt.Presented {
		user.UpdatedAt = ud.UpdatedAt.Value
	}
	if ud.Search.Presented {
		user.Search = ud.Search.Value
	}
}

type User struct {
	MongoId   *string                `json:"_id,omitempty" bson:"_id,omitempty"` // https://stackoverflow.com/a/20739427
	Id        *helper.FlexInt        `json:"id" db:"id" bson:"id,omitempty" example:"2"`
	Name      string                 `json:"name" db:"name" bson:"name,omitempty" example:"emma"`
	Password  *string                `json:"password,omitempty" db:"password" bson:"password,omitempty" example:"password"`
	Email     *string                `json:"email,omitempty" db:"email" bson:"email,omitempty" example:"xxx@example.com"`
	FirstName *string                `json:"firstName" db:"first_name" bson:"first_name,omitempty" example:"Emma"`
	LastName  *string                `json:"lastName" db:"last_name" bson:"last_name,omitempty" example:"Watson"`
	Disabled  bool                   `json:"disabled" db:"disabled" bson:"disabled,omitempty" example:"false"`
	IsOauth   bool                   `json:"isOauth" db:"is_oauth" bson:"is_oauth,omitempty" example:"false"`
	Provider  *string                `json:"provider" db:"provider" bson:"provider,omitempty" example:"google"`
	Groups    []*Group               `json:"groups"`
	CreatedAt *helper.CustomDatetime `json:"createdAt" db:"created_at"  bson:"created_at,omitempty"`
	UpdatedAt *helper.CustomDatetime `json:"updatedAt" db:"updated_at" bson:"updated_at,omitempty"`
	Search    *string                `json:"-" db:"search,omitempty" bson:"search,omitempty" example:"google"`
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

func (users Users) RowsToStruct(rows database.Rows) []*User {
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

func (users Users) GetTags(key ...string) []string {
	if len(users) == 0 {
		return []string{}
	}

	return users[0].getTags(key...)
}

func (users *Users) PrintValue() {
	for _, v := range *users {
		if v.Id != nil {
			logger.Debugf("existing --> id: %+v, v: %+v\n", *v.Id, *v)
		} else {
			logger.Debugf("new --> v: %+v\n", *v)
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
			if !slices.Contains(*database.IgnrCols, name) {
				cols = append(cols, name)
			}
		}
	}
	return cols
}
