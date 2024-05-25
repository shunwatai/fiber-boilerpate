package permissioncheck

import (
	"errors"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/groupResourceAcl"
	"testing"
)

func TestCheckPermission(t *testing.T) {
	cfg.LoadEnvVariables()
	zlog.NewZlog()

	tests := []struct {
		name              string
		reqMethod         string
		groupResourceAcls []*groupResourceAcl.GroupResourceAcl
		expectedError     error
	}{
		{
			name:      "matching permission, no error",
			reqMethod: "POST",
			groupResourceAcls: []*groupResourceAcl.GroupResourceAcl{
				{ResourceName: utils.ToPtr("users"), GroupName: utils.ToPtr("user"), PermissionType: utils.ToPtr("add")},
			},
			expectedError: nil,
		},
		{
			name:      "no matching permission, error",
			reqMethod: "GET",
			groupResourceAcls: []*groupResourceAcl.GroupResourceAcl{
				{ResourceName: utils.ToPtr("users"), GroupName: utils.ToPtr("user"), PermissionType: utils.ToPtr("edit")},
			},
			expectedError: errors.New("doesn't have permission to GET"),
		},
		{
			name:      "multiple permissions, no error",
			reqMethod: "PATCH",
			groupResourceAcls: []*groupResourceAcl.GroupResourceAcl{
				{ResourceName: utils.ToPtr("users"), GroupName: utils.ToPtr("user"), PermissionType: utils.ToPtr("edit")},
				{ResourceName: utils.ToPtr("users"), GroupName: utils.ToPtr("user"), PermissionType: utils.ToPtr("read")},
			},
			expectedError: nil,
		},
		{
			name:              "empty groupResourceAcls, error",
			reqMethod:         "DELETE",
			groupResourceAcls: []*groupResourceAcl.GroupResourceAcl{},
			expectedError:     errors.New("doesn't have permission to DELETE"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPermission(tt.reqMethod, tt.groupResourceAcls)
			if err != nil && tt.expectedError == nil {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil && tt.expectedError != nil {
				t.Errorf("expected error: %v, but got nil", tt.expectedError)
			} else if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error: %v, but got: %v", tt.expectedError, err)
				}
			}
		})
	}
}
