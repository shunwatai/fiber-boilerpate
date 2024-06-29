package user

import (
	"errors"
	"fmt"
	"golang-api-starter/internal/auth"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/groupUser"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
	ctx  *fiber.Ctx
}

func NewService(r *Repository) *Service {
	return &Service{r, nil}
}

func (s *Service) SetCtx(ctx *fiber.Ctx) {
	s.ctx = ctx
}

func (s *Service) GetLoggedInUsername() string {
	username := s.ctx.Locals("claims").(jwt.MapClaims)["username"]
	return username.(string)
}

/* this func for generate the jwt claims like the access & refresh tokens */
func GenerateUserToken(user groupUser.User, tokenType string) *jwt.Token {
	var expireTime = &jwt.NumericDate{time.Now().Add(time.Minute * 10)} // 10 mins for access token?

	env := cfg.ServerConf.Env
	if env == "local" { // if local development, set expire time to 1 year
		expireTime = &jwt.NumericDate{time.Now().Add(time.Hour * 8760)}
	}
	if tokenType == "refreshToken" {
		expireTime = &jwt.NumericDate{time.Now().Add(time.Hour * 720)} // 30 days for refresh token?
	}

	claims := &UserClaims{
		UserId: func() interface{} {
			userId := user.GetId()
			if cfg.DbConf.Driver == "mongodb" {
				return userId
			} else {
				userIdInt, _ := strconv.ParseInt(userId, 10, 64)
				return userIdInt
			}
		}(),
		Username:  user.Name,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    user.GetId(),
			ExpiresAt: expireTime,
		},
	}

	return auth.GetToken(claims)
}

func GetUserTokenResponse(user *groupUser.User) (map[string]interface{}, error) {
	accessClaims := GenerateUserToken(*user, "accessToken")
	refreshClaims := GenerateUserToken(*user, "refreshToken")

	secret := cfg.Jwt.Secret
	accessToken, accessTokenErr := accessClaims.SignedString([]byte(secret))
	refreshToken, refreshTokenErr := refreshClaims.SignedString([]byte(secret))
	if accessTokenErr != nil || refreshTokenErr != nil {
		return nil, fmt.Errorf("failed to make jwt: %+v", errors.Join(accessTokenErr, refreshTokenErr).Error())
	}

	return map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         *user,
	}, nil
}

func (s *Service) Get(queries map[string]interface{}) ([]*groupUser.User, *helper.Pagination) {
	logger.Debugf("user service get")
	users, pagination := s.repo.Get(queries)
	cascadeFields(users)

	return users, pagination
}

func (s *Service) GetById(queries map[string]interface{}) ([]*groupUser.User, error) {
	logger.Debugf("user service getById\n")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(users []*groupUser.User) ([]*groupUser.User, *helper.HttpErr) {
	logger.Debugf("user service create")
	newUserNames := []string{}
	for _, user := range users {
		newUserNames = append(newUserNames, user.Name)
	}

	// check if duplicated by "name"
	existingUsers, _ := s.repo.Get(map[string]interface{}{"name": newUserNames, "exactMatch": map[string]bool{"name": true}})

	for _, existing := range existingUsers {
		index := IndexOfDuplicatedName(users, existing)
		// new user without duplicated name
		if index < 0 {
			continue
		}

		// new user duplicated name with existing user
		if index > -1 && (users[index].Id == nil && users[index].MongoId == nil) {
			errMsg := fmt.Sprintf("user service create error: provided user name(s) %+v already exists.", newUserNames)
			return nil, &helper.HttpErr{fiber.StatusConflict, logger.Errorf(errMsg)}
		}

		// validate upsert, if id is present in JSON for updating existing user, check if id match with existing user or not
		if (users[index].Id != nil && *users[index].Id != *existing.Id) || (users[index].MongoId != nil && *users[index].MongoId != *existing.MongoId) {
			return nil, &helper.HttpErr{fiber.StatusConflict, fmt.Errorf("something went wrong, ID+Name not match with existing")}
		} 
	}

	results, err := s.repo.Create(users)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Update(users []*groupUser.User) ([]*groupUser.User, *helper.HttpErr) {
	logger.Debugf("user service update")

	// USELESS, can simply set that column as UNIQUE in DB's table.
	// check conflict of existing name
	for _, user := range users {
		if len(user.Name) == 0 {
			continue
		}
		conflicts, _ := s.repo.Get(map[string]interface{}{
			"name": user.Name,
			"exactMatch": map[string]bool{
				"name": true,
			},
		})
		if len(conflicts) > 0 && conflicts[0].GetId() != user.GetId() {
			httpErr := &helper.HttpErr{
				Code: fiber.StatusConflict,
				Err:  fmt.Errorf("%+v is already existed, please try another name.", user.Name),
			}
			return nil, httpErr
		}
	}

	results, err := s.repo.Update(users)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*groupUser.User, error) {
	logger.Debugf("user service delete")
	records := []*groupUser.User{}
	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ = s.repo.Get(getByIdsCondition)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}

func (s *Service) Login(user *groupUser.User) (map[string]interface{}, *helper.HttpErr) {
	logger.Debugf("user service login")

	results, _ := s.repo.Get(map[string]interface{}{
		"name": user.Name,
		"exactMatch": map[string]bool{
			"name": true,
		},
	})
	if len(results) == 0 {
		return nil, &helper.HttpErr{fiber.StatusNotFound, fmt.Errorf("user not exists...")}
	}
	logger.Debugf("results?? %+v", results)

	if !user.IsOauth {
		var checkPassword = func(hashedPwd string, plainPwd string) bool {
			if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd)); err != nil {
				return false
			}

			return true
		}

		match := checkPassword(*results[0].Password, *user.Password)

		if !match {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, fmt.Errorf("password not match...")}
		}
	}

	sanitise(results)

	if userTokenResponse, err := GetUserTokenResponse(results[0]); err != nil {
		msg := fmt.Sprintf("failed to refresh token: %+v", err)
		logger.Errorf(msg)
		return nil, &helper.HttpErr{fiber.StatusInternalServerError, errors.New(msg)}
	} else {
		return userTokenResponse, nil
	}
}

func (s *Service) Refresh(user *groupUser.User) (map[string]interface{}, *helper.HttpErr) {
	logger.Debugf("user service refresh")

	results := []*groupUser.User{}
	getByIdsCondition := database.GetIdsMapCondition(nil, []string{user.GetId()})
	results, _ = s.repo.Get(getByIdsCondition)
	if len(results) == 0 {
		return nil, &helper.HttpErr{fiber.StatusNotFound, fmt.Errorf("user not exists... failed to refresh, please try login again")}
	}

	sanitise(results)
	if userTokenResponse, err := GetUserTokenResponse(results[0]); err != nil {
		return nil, &helper.HttpErr{fiber.StatusNotFound, fmt.Errorf("failed to refresh token: %+v", err.Error())}
	} else {
		return userTokenResponse, nil
	}
}

func IndexOfDuplicatedName(users groupUser.Users, existingUser *groupUser.User) int {
	for i, u := range users {
		if u.Name == existingUser.Name {
			return i
		}
	}
	return -1
}
