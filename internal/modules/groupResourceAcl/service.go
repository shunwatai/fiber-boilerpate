package groupResourceAcl

import (
	"fmt"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
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
func (s *Service) checkUpdateNonExistRecord(groupResourceAcl *GroupResourceAcl) error {
	conditions := map[string]interface{}{}
	conditions["id"] = groupResourceAcl.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if groupResourceAcl.CreatedAt == nil {
		groupResourceAcl.CreatedAt = existing[0].CreatedAt
	}

	return nil
}

func (s *Service) GetGroupIdMap(gras []*GroupResourceAcl) map[string][]*GroupResourceAcl {
	groupResourceAclsMap := map[string][]*GroupResourceAcl{}
	for _, gra := range gras {
		groupResourceAclsMap[gra.GetGroupId()] = append(groupResourceAclsMap[gra.GetGroupId()], gra)
	}
	return groupResourceAclsMap
}

func (s *Service) Get(queries map[string]interface{}) ([]*GroupResourceAcl, *helper.Pagination) {
	logger.Debugf("groupResourceAcl service get")
	return s.repo.Get(queries)
}

func (s *Service) GetById(queries map[string]interface{}) ([]*GroupResourceAcl, error) {
	logger.Debugf("groupResourceAcl service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(groupResourceAcls []*GroupResourceAcl) ([]*GroupResourceAcl, *helper.HttpErr) {
	logger.Debugf("groupResourceAcl service create")
  /*
	// use the claims for mark the "createdBy/updatedBy" in database
	claims := s.ctx.Locals("claims").(jwt.MapClaims)
	logger.Debugf("req by userId: %+v, username: %+v", claims["userId"], claims["username"])
	for _, groupResourceAcl := range groupResourceAcls {
		if groupResourceAcl.UserId == nil {
			groupResourceAcl.UserId = claims["userId"]
		}
		if validErr := helper.ValidateStruct(*groupResourceAcl); validErr != nil {
			return nil, &helper.HttpErr{fiber.StatusUnprocessableEntity, validErr}
		}
	}
  */

	results, err := s.repo.Create(groupResourceAcls)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Update(groupResourceAcls []*GroupResourceAcl) ([]*GroupResourceAcl, *helper.HttpErr) {
	logger.Debugf("groupResourceAcl service update")
	for _, groupResourceAcl := range groupResourceAcls {
		if err := s.checkUpdateNonExistRecord(groupResourceAcl); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}
	results, err := s.repo.Update(groupResourceAcls)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*GroupResourceAcl, error) {
	logger.Debugf("groupResourceAcl service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}
