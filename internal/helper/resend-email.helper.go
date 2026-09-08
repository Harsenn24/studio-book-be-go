package helper

import (
	"context"
	"fmt"
	"os"
	emailLoggerRepo "studio-book-be-go/internal/repository/email_logger"
	userRepo "studio-book-be-go/internal/repository/user"
	"gorm.io/gorm"
)

type ResendEmailResult struct {
	Status  bool
	Message string
}

type ResendEmailHelper struct {
	emailLoggerRepo emailLoggerRepo.EmailLoggerRepository
	db              *gorm.DB
}

func NewResendEmailHelper(emailLoggerRepo emailLoggerRepo.EmailLoggerRepository, db *gorm.DB) *ResendEmailHelper {
	return &ResendEmailHelper{
		db:              db,
		emailLoggerRepo: emailLoggerRepo,
	}
}

func (r *ResendEmailHelper) ReSendEmail(
	ctx context.Context,
	userData userRepo.User,
	email string,
	role string,
) (*ResendEmailResult, error) {

	emailLogs, err := r.emailLoggerRepo.FindAllByEmailAndType(email, "verified-user")
	if err != nil {
		return nil, err
	}

	if len(emailLogs) > 3 {
		return &ResendEmailResult{
			Status:  false,
			Message: "Email has been limited",
		}, nil
	}

	generateToken, err := GenerateToken(userData.UUID, userData.Email, role)
	if err != nil {
		return nil, err
	}

	nodeEnv := os.Getenv("NODE_ENV")
	clientURL := os.Getenv("CLIENT_URL_DEV")
	if nodeEnv == "production" {
		clientURL = os.Getenv("CLIENT_URL")
	}

	payloadEmail := map[string]interface{}{
		"name": userData.Name,
		"link": fmt.Sprintf("%s/verify/%s/%s?token=%s", clientURL, role, userData.UUID, generateToken),
	}

	emailHelper := NewSendEmailHelper(r.emailLoggerRepo, r.db)
	resultSendEmail, err := emailHelper.SendEmail(ctx, email, "Email verification", "verified-user", payloadEmail, userData.ID)

	if err != nil {
		return nil, err
	}

	if len(resultSendEmail.Rejected) > 0 {
		return &ResendEmailResult{
			Status:  false,
			Message: "Failed to send email",
		}, nil
	}

	return &ResendEmailResult{
		Status:  true,
		Message: "Verification email re-sent",
	}, nil
}
