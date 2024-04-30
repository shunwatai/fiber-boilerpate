package passwordReset

import (
	"fmt"
	"golang-api-starter/internal/auth"
	"golang-api-starter/internal/database"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	"golang-api-starter/internal/modules/user"
	"golang-api-starter/internal/notification/email"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
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
func (s *Service) checkUpdateNonExistRecord(passwordReset *PasswordReset) error {
	conditions := map[string]interface{}{}
	conditions["id"] = passwordReset.GetId()

	existing, _ := s.repo.Get(conditions)
	if len(existing) == 0 {
		respCode = fiber.StatusNotFound
		return logger.Errorf("cannot update non-existing records...")
	} else if passwordReset.CreatedAt == nil {
		passwordReset.CreatedAt = existing[0].CreatedAt
	}

	return nil
}

func (s *Service) Get(queries map[string]interface{}) ([]*PasswordReset, *helper.Pagination) {
	logger.Debugf("passwordReset service get")
	return s.repo.Get(queries)
}

func (s *Service) GetById(queries map[string]interface{}) ([]*PasswordReset, error) {
	logger.Debugf("passwordReset service getById")

	records, _ := s.repo.Get(queries)
	if len(records) == 0 {
		return nil, fmt.Errorf("%s with id: %s not found", tableName, queries["id"])
	}
	return records, nil
}

func (s *Service) Create(passwordResets []*PasswordReset) ([]*PasswordReset, *helper.HttpErr) {
	logger.Debugf("passwordReset service create")
	userIdMap := map[string]PasswordReset{}

	for _, pr := range passwordResets {
		// get the user by passwordReset.Email
		// proceed only record found, else throw err 404
		users, _ := user.Srvc.Get(map[string]interface{}{
			"email":      pr.Email,
			"exactMatch": map[string]bool{"email": true},
		})
		if len(users) == 0 {
			return nil, &helper.HttpErr{fiber.StatusNotFound, fmt.Errorf("Fail to find the user that match with email: %s", pr.Email)}
		}

		// construct the record for insert into db here
		// logger.Debugf("users[0].GetId(): %+v", users[0].GetId())
		pr.UserId = users[0].GetId()
		expiryDate := time.Now().Add(time.Hour * 24)
		pr.ExpiryDate = &helper.CustomDatetime{utils.ToPtr(expiryDate), utils.ToPtr(time.RFC3339)}
		// logger.Debugf("pr.ExpiryDate.Time: %+v", *pr.ExpiryDate.Time)
		pr.Token = utils.GetRandString(4)
		if hashBytes, err := bcrypt.GenerateFromPassword([]byte(pr.Token), bcrypt.MinCost); err != nil {
			return nil, &helper.HttpErr{fiber.StatusNotFound, fmt.Errorf("Fail to hash token")}
		} else {
			pr.TokenHash = utils.ToPtr(string(hashBytes))
		}

		userIdMap[pr.UserId.(string)] = *pr
	}

	// revoke all previous/existing records by setting IsUsed to true before add the new record
	existingRecords, _ := s.repo.Get(map[string]interface{}{"user_id": passwordResets[0].UserId})
	for _, record := range existingRecords {
		record.IsUsed = true
	}
	s.repo.Update(existingRecords)

	results, err := s.repo.Create(passwordResets)
	if err == nil { // if create success, send email
		err = s.sendResetPasswordEmail(passwordResets)
	}

	// useless, just want to map the email & token back in response
	for _, r := range results {
		pr := userIdMap[r.GetUserId()]
		r.Email = pr.Email
		r.Token = pr.Token
	}

	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) sendResetPasswordEmail(passwordResets []*PasswordReset) error {
	for _, pr := range passwordResets {
		resetLink := fmt.Sprintf("%s/password-resets?token=%s&userId=%s&email=%s", cfg.GetServerUrl(), pr.Token, pr.UserId, pr.Email)

		tmplFiles := []string{"web/template/reset-password/reset-email.gohtml"}

		emailInfo := email.EmailInfo{
			To: []string{pr.Email},
			MsgMeta: map[string]interface{}{
				"subject":          "reset password",
				"resetPasswordUrl": resetLink,
			},
			Template: template.Must(template.ParseFiles(tmplFiles...)),
		}

		if err := email.TemplateEmail(emailInfo); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Update(passwordResets []*PasswordReset) ([]*PasswordReset, *helper.HttpErr) {
	logger.Debugf("passwordReset service update")
	for _, passwordReset := range passwordResets {
		if err := s.checkUpdateNonExistRecord(passwordReset); err !=nil{
			return nil, &helper.HttpErr{fiber.StatusInternalServerError, err}
		}
	}
	results, err := s.repo.Update(passwordResets)
	return results, &helper.HttpErr{fiber.StatusInternalServerError, err}
}

func (s *Service) Delete(ids []string) ([]*PasswordReset, error) {
	logger.Debugf("passwordReset service delete")

	getByIdsCondition := database.GetIdsMapCondition(nil, ids)
	records, _ := s.repo.Get(getByIdsCondition)
	logger.Debugf("records: %+v", records)
	if len(records) == 0 {
		return nil, fmt.Errorf("failed to delete, %s with id: %+v not found", tableName, ids)
	}

	return records, s.repo.Delete(ids)
}

func GetResetJwtToken(passwordReset *PasswordReset) (string, error) {
	var resetToken string
	var resetTokenErr error
	secret := cfg.Jwt.Secret
	resetClaims := GenerateResetToken(*passwordReset, "accessToken")
	if resetToken, resetTokenErr = resetClaims.SignedString([]byte(secret)); resetTokenErr != nil {
		return "", logger.Errorf("failed to make jwt: %+v", resetTokenErr.Error())
	}
	return resetToken, nil
}

func GenerateResetToken(passwordReset PasswordReset, tokenType string) *jwt.Token {
	// env := cfg.ServerConf.Env
	var expireTime = &jwt.NumericDate{time.Now().Add(time.Minute * 10)} // 10 mins for access token?

	claims := &ResetClaims{
		UserId:    passwordReset.UserId,
		Email:     passwordReset.Email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "admin",
			ExpiresAt: expireTime,
		},
	}

	return auth.GetToken(claims)
}
