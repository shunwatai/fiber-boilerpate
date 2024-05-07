package groupUser

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
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
func (s *Service) checkUpdateNonExistRecord(groupUser *GroupUser) error {
	conditions := map[string]interface{}{}
	conditions["id"] = groupUser.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if groupUser.CreatedAt == nil {
		groupUser.CreatedAt = existing[0].CreatedAt
	}

	return nil
}

func (s *Service) GetGroupIdMap(gus []*GroupUser) map[string][]*GroupUser {
	groupUsersMap := map[string][]*GroupUser{}
	for _, gu := range gus {
		groupUsersMap[gu.GetGroupId()] = append(groupUsersMap[gu.GetGroupId()], gu)
	}
	return groupUsersMap
}

func (s *Service) Get(queries map[string]interface{}) ([]*GroupUser, *helper.Pagination) {
	logger.Debugf("groupUser service get")
	return s.repo.Get(queries)
}

func (s *Service) GetById(queries map[string]interface{}) ([]*GroupUser, error) {
	logger.Debugf("groupUser service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(groupUsers []*GroupUser) ([]*GroupUser, *helper.HttpErr) {
	logger.Debugf("groupUser service create")
	/*
		// use the claims for mark the "createdBy/updatedBy" in database
		claims := s.ctx.Locals("claims").(jwt.MapClaims)
		logger.Debugf("req by userId: %+v, username: %+v", claims["userId"], claims["username"])
		for _, groupUser := range groupUsers {
			if groupUser.UserId == nil {
				groupUser.UserId = claims["userId"]
			}
			if validErr := helper.ValidateStruct(*groupUser); validErr != nil {
				return nil, &helper.HttpErr{fiber.StatusUnprocessableEntity, validErr}
			}
		}
	*/

	results, err := s.repo.Create(groupUsers)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Update(groupUsers []*GroupUser) ([]*GroupUser, *helper.HttpErr) {
	logger.Debugf("groupUser service update")
	for _, groupUser := range groupUsers {
		if err := s.checkUpdateNonExistRecord(groupUser); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}
	results, err := s.repo.Update(groupUsers)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*GroupUser, error) {
	logger.Debugf("groupUser service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}
