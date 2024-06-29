package group

import (
	"fmt"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/modules/groupResourceAcl"
	"golang-api-starter/internal/modules/groupUser"
	"golang-api-starter/internal/modules/permissionType"
	"golang-api-starter/internal/modules/resource"

	"github.com/gofiber/fiber/v2"
)

type Service struct {
	repo *Repository
	ctx  *fiber.Ctx
}

func NewService(r *Repository) *Service {
	return &Service{r, nil}
}

// checkUpdateNonExistRecord for the "update" function to remain the createdAt value without accidental alter the createdAt
// it may slow, should follow user/service.go's Update to fetch all records at once to reduce db fetching
func (s *Service) checkUpdateNonExistRecord(group *groupUser.Group) error {
	conditions := map[string]interface{}{}
	conditions["id"] = group.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if group.CreatedAt == nil {
		group.CreatedAt = existing[0].CreatedAt
	}

	if len(group.Type) == 0 {
		group.Type = existing[0].Type
	}

	return nil
}

// isDuplicated check for duplicated name in DB
// this function specifically made for Mysql because of its ON DUPLICATE KEY UPDATE can't ignore the UNIQUE index...
func (s *Service) isDuplicated(group *groupUser.Group) error {
	conditions := map[string]interface{}{}
	conditions["name"] = group.Name

	existing, _ := s.repo.Get(conditions)
	// no duplicated, return
	if len(existing) == 0 {
		return nil
	}

	if existing[0].Name == group.Name {
		respCode = fiber.StatusConflict
		return logger.Errorf(fmt.Sprintf("group with name:%s already exists...", group.Name))
	}

	return nil
}

func (s *Service) Get(queries map[string]interface{}) ([]*groupUser.Group, *helper.Pagination) {
	logger.Debugf("group service get")
	groups, pagination := s.repo.Get(queries)
	cascadeFields(groups)

	return groups, pagination
}

func (s *Service) GetById(queries map[string]interface{}) ([]*groupUser.Group, error) {
	logger.Debugf("group service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	cascadeFields(records)

	return records, nil
}

func (s *Service) Create(groups []*groupUser.Group) ([]*groupUser.Group, *helper.HttpErr) {
	logger.Debugf("group service create")
	// can remove this check if NOT using mariadb.
	// pg,sqlite,monogo can throu error with ON DUPLICATE KEY UPDATE even with the 'name' as UNIQUE
	for _, group := range groups {
		if err := s.isDuplicated(group); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}

	results, err := s.repo.Create(groups)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Update(groups []*groupUser.Group) ([]*groupUser.Group, *helper.HttpErr) {
	logger.Debugf("group service update")

	for _, group := range groups {
		if err := s.checkUpdateNonExistRecord(group); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}

		// only update users:[] & permissions:[] when PATch single record.
		// it is because of frontend is difficult to pass the validation of these 2 fields when batch updating "disabled"  
		if len(groups) == 1 {
			updateGroupUsers(group)
			updateGroupResourceAcls(group)
		}
	}
	results, err := s.repo.Update(groups)

	cascadeFields(results)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*groupUser.Group, error) {
	logger.Debugf("group service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}

func updateGroupUsers(group *groupUser.Group) {
	// remove all existing groupUsers records
	existingGroupUsers, _ := groupUser.Srvc.Get(map[string]interface{}{"group_id": group.GetId()})
	existingGroupUsersIds := []string{}
	for _, gu := range existingGroupUsers {
		existingGroupUsersIds = append(existingGroupUsersIds, gu.GetId())
	}
	if len(existingGroupUsersIds) > 0 {
		groupUser.Srvc.Delete(existingGroupUsersIds)
	}
	// update groupUsers table
	if len(group.Users) > 0 {
		groupUsers := []*groupUser.GroupUser{}
		for _, u := range group.Users {
			groupUsers = append(groupUsers, &groupUser.GroupUser{GroupId: group.GetId(), UserId: u.GetId()})
		}

		groupUser.Srvc.Create(groupUsers)
	}
}

func updateGroupResourceAcls(group *groupUser.Group) {
	resources, _ := resource.Srvc.Get(map[string]interface{}{})
	resourceNameMap := resource.Resources(resources).GetNameMap()
	permissionTypes, _ := permissionType.Srvc.Get(map[string]interface{}{})
	permissionTypeNameMap := permissionType.PermissionTypes(permissionTypes).GetNameMap()
	// logger.Debugf("resourceNameMap: %+v, permissionTypeNameMap: %+v", resourceNameMap, permissionTypeNameMap)

	// remove all existing groupResourceAcls records
	existingGroupResourceAcls, _ := groupResourceAcl.Srvc.Get(map[string]interface{}{"group_id": group.GetId()})
	existingGroupResourceAclsIds := []string{}
	for _, gu := range existingGroupResourceAcls {
		existingGroupResourceAclsIds = append(existingGroupResourceAclsIds, gu.GetId())
	}
	if len(existingGroupResourceAclsIds) > 0 {
		groupResourceAcl.Srvc.Delete(existingGroupResourceAclsIds)
	}
	// update groupResourceAcls table
	if len(group.Permissions) > 0 {
		groupResourceAcls := []*groupResourceAcl.GroupResourceAcl{}
		for _, perm := range group.Permissions {
			if resourceNameMap[*perm.ResourceName] == nil ||
				permissionTypeNameMap[*perm.PermissionType] == nil {
				continue
			}
			resourceId := resourceNameMap[*perm.ResourceName].GetId()
			permissionTypeId := permissionTypeNameMap[*perm.PermissionType].GetId()
			groupResourceAcls = append(groupResourceAcls, &groupResourceAcl.GroupResourceAcl{
				GroupId:          group.GetId(),
				ResourceId:       resourceId,
				PermissionTypeId: permissionTypeId,
			})
		}

		groupResourceAcl.Srvc.Create(groupResourceAcls)
	}
}
