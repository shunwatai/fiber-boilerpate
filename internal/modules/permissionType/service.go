package permissionType

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
func (s *Service) checkUpdateNonExistRecord(permissionType *PermissionType) error {
	conditions := map[string]interface{}{}
	conditions["id"] = permissionType.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if permissionType.CreatedAt == nil {
		permissionType.CreatedAt = existing[0].CreatedAt
	}

	return nil
}

func (s *Service) Get(queries map[string]interface{}) ([]*PermissionType, *helper.Pagination) {
	logger.Debugf("permissionType service get")
	return s.repo.Get(queries)
}

func (s *Service) GetById(queries map[string]interface{}) ([]*PermissionType, error) {
	logger.Debugf("permissionType service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(permissionTypes []*PermissionType) ([]*PermissionType, *helper.HttpErr) {
	logger.Debugf("permissionType service create")
  /*
	// use the claims for mark the "createdBy/updatedBy" in database
	claims := s.ctx.Locals("claims").(jwt.MapClaims)
	logger.Debugf("req by userId: %+v, username: %+v", claims["userId"], claims["username"])
	for _, permissionType := range permissionTypes {
		if permissionType.UserId == nil {
			permissionType.UserId = claims["userId"]
		}
		if validErr := helper.ValidateStruct(*permissionType); validErr != nil {
			return nil, &helper.HttpErr{fiber.StatusUnprocessableEntity, validErr}
		}
	}
  */

	results, err := s.repo.Create(permissionTypes)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Update(permissionTypes []*PermissionType) ([]*PermissionType, *helper.HttpErr) {
	logger.Debugf("permissionType service update")
	for _, permissionType := range permissionTypes {
		if err := s.checkUpdateNonExistRecord(permissionType); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}
	results, err := s.repo.Update(permissionTypes)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*PermissionType, error) {
	logger.Debugf("permissionType service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}
