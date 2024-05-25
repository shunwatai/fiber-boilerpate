package groupUser

import (
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"

	"golang.org/x/exp/maps"
)

type Repository struct {
	db        database.IDatabase
	UserRepo  IUserRepository
	GroupRepo IGroupRepository
}

type IUserRepository interface {
	Get(queries map[string]interface{}) ([]*User, *helper.Pagination)
	GetIdMap(users Users) map[string]*User
}
type IGroupRepository interface {
	Get(queries map[string]interface{}) ([]*Group, *helper.Pagination)
	GetIdMap(groups Groups) map[string]*Group
}

func NewRepository(db database.IDatabase) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetGroupIdMap(gus []*GroupUser) map[string][]*GroupUser {
	groupUsersMap := map[string][]*GroupUser{}
	for _, gu := range gus {
		groupUsersMap[gu.GetGroupId()] = append(groupUsersMap[gu.GetGroupId()], gu)
	}
	return groupUsersMap
}

func (r *Repository) GetUserIdMap(gus []*GroupUser) map[string][]*GroupUser {
	groupUsersMap := map[string][]*GroupUser{}
	for _, gu := range gus {
		groupUsersMap[gu.GetUserId()] = append(groupUsersMap[gu.GetUserId()], gu)
	}
	return groupUsersMap
}

// cascadeFields for joining other module, see the example in internal/modules/todo/repository.go
func cascadeFields(groupUsers GroupUsers) {
	if len(groupUsers) == 0 {
		return
	}

	// map user & group
	var (
		userIds  []string
		groupIds []string
	)
	// get all userIds & groupIds
	for _, groupUser := range groupUsers {
		userIds = append(userIds, groupUser.GetUserId())
		groupIds = append(groupIds, groupUser.GetGroupId())
	}

	if len(userIds) > 0 {
		// get user by userIds
		condition := database.GetIdsMapCondition(utils.ToPtr("user_id"), userIds)
		users, _ := Repo.UserRepo.Get(condition)
		// get the map[userId]User
		userMap := Repo.UserRepo.GetIdMap(users)

		for _, groupUser := range groupUsers {
			groupUser.User = new(User)
			// take out the document by documentId in map and assign
			user := userMap[groupUser.GetUserId()]
			groupUser.User = user
		}
	}

	if len(groupIds) > 0 {
		// get group by groupIds
		condition := database.GetIdsMapCondition(utils.ToPtr("group_id"), groupIds)
		groups, _ := Repo.GroupRepo.Get(condition)
		// get the map[groupId]group
		groupMap := Repo.GroupRepo.GetIdMap(groups)

		for _, gropuUser := range groupUsers {
			gropuUser.Group = new(Group)
			// take out the document by documentId in map and assign
			group := groupMap[gropuUser.GetGroupId()]
			gropuUser.Group = group
		}
	}
}

func (r *Repository) Get(queries map[string]interface{}) ([]*GroupUser, *helper.Pagination) {
	logger.Debugf("groupUser repo get")
	defaultExactMatch := map[string]bool{
		"id":  true,
		"_id": true,
		//"done": true, // bool match needs exact match, param can be 0(false) & 1(true)
	}
	if queries["exactMatch"] != nil {
		maps.Copy(queries["exactMatch"].(map[string]bool), defaultExactMatch)
	} else {
		queries["exactMatch"] = defaultExactMatch
	}

	queries["columns"] = GroupUser{}.getTags()
	rows, pagination := r.db.Select(queries)

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	// records.printValue()

	cascadeFields(records)

	return records, pagination
}

func (r *Repository) Create(groupUsers []*GroupUser) ([]*GroupUser, error) {
	for _, groupUser := range groupUsers {
		logger.Debugf("groupUser repo add: %+v", groupUser)
	}
	database.SetIgnoredCols("search")
	defer database.SetIgnoredCols()
	rows, err := r.db.Save(GroupUsers(groupUsers))

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Update(groupUsers []*GroupUser) ([]*GroupUser, error) {
	logger.Debugf("groupUser repo update")
	rows, err := r.db.Save(GroupUsers(groupUsers))

	var records GroupUsers
	if rows != nil {
		records = records.rowsToStruct(rows)
	}
	records.printValue()

	return records, err
}

func (r *Repository) Delete(ids []string) error {
	logger.Debugf("groupUser repo delete")
	err := r.db.Delete(ids)
	if err != nil {
		return err
	}

	return nil
}
