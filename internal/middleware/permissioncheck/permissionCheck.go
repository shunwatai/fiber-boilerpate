package permissioncheck

import (
	"fmt"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/groupResourceAcl"
	"golang-api-starter/internal/modules/groupUser"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var cfg = config.Cfg

type PermissionChecker struct{}

// CheckAccess is the middleware for checking the access permission by the records of groupResourceAcls in DB.
// The arg resourceName is mapped by the resourceId to the resource table in DB.
func (pc *PermissionChecker) CheckAccess(resourceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fctx := &helper.FiberCtx{Fctx: c}
		reqMethod := c.Method()
		// log.Printf("1reqBody: %+v, %+v \n", len(string(bodyBytes)), string(bodyBytes))
		claims := c.Locals("claims").(jwt.MapClaims)

		var userId string
		if cfg.DbConf.Driver == "mongodb" {
			userId = claims["userId"].(string)
		} else {
			userId = strconv.Itoa(int(claims["userId"].(float64)))
		}
		logger.Debugf("userId???? %+v\n", userId)

		// get groupUsers by userId
		groupUsers, _ := groupUser.Srvc.Get(map[string]interface{}{"user_id": userId})
		if len(groupUsers) == 0 {
			return fctx.ErrResponse(fiber.StatusForbidden, logger.Errorf("userId: %+v doesn't belong to any group", userId))
		}

		groupIds := []string{}
		for _, gu := range groupUsers {
			// if user in admin group, skip permission check
			if gu.IsAdmin() {
				return c.Next()
			}
			groupIds = append(groupIds, gu.GetGroupId())
		}

		// get groupResourceAcls by groupIds & resourceName
		groupResourceAcls, _ := groupResourceAcl.Srvc.Get(map[string]interface{}{
			"group_id":      groupIds,
			"resource_name": resourceName,
		})

		if err := checkPermission(reqMethod, groupResourceAcls); err != nil {
			return fctx.ErrResponse(fiber.StatusForbidden, fmt.Errorf("userId: %+v in groupId: %+v %+v to %+v", userId, groupIds, err.Error(), resourceName))
		}

		return c.Next()
	}
}

// checkPermission loop all groupResourceAcls to check for any permissionType match for the corresponding resquestMethod.
// The groupResourceAcls are already selected by user's groupIds and against the target resourceName(module)
func checkPermission(reqMethod string, groupResourceAcls []*groupResourceAcl.GroupResourceAcl) error {
	// logger.Debugf("req method???? %+v\n", reqMethod)
	hasPermissions := groupResourceAcl.GroupResourceAcls{}

	methodToPermType := map[string]string{
		"GET":    "read",
		"POST":   "add",
		"PATCH":  "edit",
		"PUT":    "edit",
		"DELETE": "delete",
	}

	for _, gra := range groupResourceAcls {
		logger.Debugf("gra resName: %+v, gra permType: %+v\n", *gra.ResourceName, *gra.PermissionType)
		// check if there is any *gra.PermissionType matches with request method
		if permType, ok := methodToPermType[reqMethod]; ok && *gra.PermissionType == permType {
			hasPermissions = append(hasPermissions, gra)
		}
	}

	if len(hasPermissions) == 0 { // no permission, respond err
		return logger.Errorf("doesn't have permission to %+v", reqMethod)
	}

	return nil
}
