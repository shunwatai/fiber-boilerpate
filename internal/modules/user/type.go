package user

import (
	"github.com/golang-jwt/jwt/v4"
)

type UserClaims struct {
	UserId    interface{} `json:"userId"`
	Username  string      `json:"username"`
	TokenType string      `json:"tokenType"`
	jwt.RegisteredClaims
}

// other user related types are grouped in internal/modules/groupUser/user_type.go
