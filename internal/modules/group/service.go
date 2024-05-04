package group

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
func (s *Service) checkUpdateNonExistRecord(group *Group) error {
	conditions := map[string]interface{}{}
	conditions["id"] = group.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if group.CreatedAt == nil {
		group.CreatedAt = existing[0].CreatedAt
	}

	return nil
}

// isDuplicated check for duplicated name in DB
// this function specifically made for Mysql because of its ON DUPLICATE KEY UPDATE can't ignore the UNIQUE index...
func (s *Service) isDuplicated(group *Group) error {
	conditions := map[string]interface{}{}
	conditions["name"] = group.Name

	existing, _ := s.repo.Get(conditions)
	// no duplicated, return
	if len(existing) == 0 { 
		return nil
	}

	if existing[0].Name == group.Name {
		respCode = fiber.StatusConflict
		return logger.Errorf(fmt.Sprintf("group wiht name:%s already exists...", group.Name))
	}

	return nil
}

func (s *Service) Get(queries map[string]interface{}) ([]*Group, *helper.Pagination) {
	logger.Debugf("group service get")
	return s.repo.Get(queries)
}

func (s *Service) GetById(queries map[string]interface{}) ([]*Group, error) {
	logger.Debugf("group service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(groups []*Group) ([]*Group, *helper.HttpErr) {
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

func (s *Service) Update(groups []*Group) ([]*Group, *helper.HttpErr) {
	logger.Debugf("group service update")
	for _, group := range groups {
		if err := s.checkUpdateNonExistRecord(group); err != nil {
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}
	results, err := s.repo.Update(groups)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*Group, error) {
	logger.Debugf("group service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}
