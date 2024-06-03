package jwtcheck

import (
	"errors"
	"golang-api-starter/internal/auth"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type JwtChecker struct{}

/*
* CheckJwt is a middleware for checking the jwt in both cookie & header
* it will first check the cookie, if failed then check the header
 */
func (jc *JwtChecker) CheckJwt(ignorePaths ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// logger.Debugf("middleware checking jwt in header.....")
		url := string(c.Request().URI().Path())
		// logger.Debugf("jwt check: %+v, %+v", url, slices.Contains(ignorePaths, url))
		if slices.Contains(ignorePaths, url) {
			return c.Next()
		}

		requestHeader := c.GetReqHeaders()
		isHtml := strings.Contains(requestHeader["Accept"][0], "text/html")

		var (
			claims jwt.MapClaims
			errStr []string
		)

		claims, err := GetTokenFromCookie(c)
		if err == nil {
			c.Locals("claims", claims)
			return c.Next()
		} else {
			errStr = append(errStr, err.Error())
		}

		claims, err = GetTokenFromHeader(c)
		if err != nil {
			errStr = append(errStr, err.Error())
		}

		if err != nil {
			if isHtml {
				c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
				return c.Redirect("/error", fiber.StatusTemporaryRedirect)
			}

			return c.
				Status(fiber.StatusUnauthorized).
				JSON(map[string]interface{}{"message": errors.Join(errors.New(strings.Join(errStr, ". ")), errors.New("failed to get the jwt from both cookie & header")).Error()})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

func GetTokenFromHeader(ctx *fiber.Ctx) (jwt.MapClaims, error) {
	accessToken := ctx.Get("Authorization")
	if len(accessToken) == 0 {
		return nil, logger.Errorf("Authorization isn't present in header")
	}

	claims, err := auth.ParseJwt(accessToken)
	if err != nil {
		return nil, logger.Errorf("failed to parse token from header, err: %+v", err)
	}

	return claims, nil
}

func GetTokenFromCookie(ctx *fiber.Ctx) (jwt.MapClaims, error) {
	jwt := ctx.Cookies("accessToken")
	if len(jwt) == 0 {
		return nil, logger.Errorf("cookie['accessToken'] isn't present")
	}

	accessToken := "Bearer " + jwt
	logger.Debugf("%s\n", accessToken)

	claims, err := auth.ParseJwt(accessToken)
	if err != nil {
		return nil, logger.Errorf("failed to parse token from cookie, err: %+v", err)
	}

	return claims, nil
}
