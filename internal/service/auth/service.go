package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"studio-book-be-go/internal/helper"
	emailLoggerRepo "studio-book-be-go/internal/repository/email_logger"
	user "studio-book-be-go/internal/repository/user"
	user_logger "studio-book-be-go/internal/repository/user_logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, input LoginRequest, role string) (LoginResponse, error)
	Register(ctx context.Context, input RegisterRequest, role string) (RegisterResponse, error)
	CheckUser(ctx context.Context, input CheckUserRequest, role string) (CheckUserResponse, error)
	VerifyEmail(ctx context.Context, input VerifyEmailRequest) (VerifyEmailResponse, error)
}

type authServiceImpl struct {
	userRepo        user.UserRepository
	userLoggerRepo  user_logger.UserLoggerRepository
	emailLoggerRepo emailLoggerRepo.EmailLoggerRepository
	db              *gorm.DB
}

func NewAuthService(
	db *gorm.DB,
	userRepo user.UserRepository,
	userLoggerRepo user_logger.UserLoggerRepository,
	emailLoggerRepo emailLoggerRepo.EmailLoggerRepository,
) AuthService {
	return &authServiceImpl{
		db:              db,
		userRepo:        userRepo,
		userLoggerRepo:  userLoggerRepo,
		emailLoggerRepo: emailLoggerRepo,
	}
}

func (s *authServiceImpl) VerifyEmail(ctx context.Context, input VerifyEmailRequest) (VerifyEmailResponse, error) {

	decodeToken, err := helper.DecodeToken(input.TokenVerify)
	if err != nil {
		return VerifyEmailResponse{}, errors.New("verify_email_failed [invalid_token]")
	}

	if(decodeToken.ExpiresAt.Before(time.Now())) {
		return VerifyEmailResponse{}, errors.New("verify_email_failed [expired_token]")
	}

	filterFindOneBy := user.User{
		UUID: decodeToken.UUID,
		Role: input.Role,
	}

	userData, err := s.userRepo.FindOneBy(ctx, filterFindOneBy)
	if err != nil {
		return VerifyEmailResponse{}, errors.New("verify_email_failed [user_not_found]")
	}

	if userData.EmailVerifiedAt != nil {
		return VerifyEmailResponse{}, errors.New("verify_email_failed [email_verified]")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()

		_, err := s.userRepo.UpdateByID(user.User{ID: userData.ID, EmailVerifiedAt: &now}, tx)
		if err != nil {
			return err
		}
		return nil
	})

	generateToken, err := helper.GenerateToken(userData.UUID, userData.Email, input.Role)
	if err != nil {
		return VerifyEmailResponse{}, errors.New("verify_email_failed [generate_token]")
	}


	return VerifyEmailResponse{
		Token: generateToken,
	}, nil
}


func (s *authServiceImpl) CheckUser(ctx context.Context, input CheckUserRequest, role string) (CheckUserResponse, error) {
	filterFindOneBy := user.User{
		Email: input.Email,
		Role:  role,
	}

	checkUser, err := s.userRepo.FindOneBy(ctx, filterFindOneBy)
	if err != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [user_not_found]")
	}

	if checkUser.EmailVerifiedAt != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [email_verified]")
	}

	emailLogs, err := s.emailLoggerRepo.FindAllByEmailAndType(input.Email, "verified-user")
	if err != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [email_logs]")
	}

	if len(emailLogs) >= 3 {
		return CheckUserResponse{}, errors.New("check_user_failed [email_logs_limit]")
	}

	generateToken, err := helper.GenerateToken(checkUser.UUID, checkUser.Email, role)
	if err != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [generate_token]")
	}

	baseURL := os.Getenv("URL_DEV")
	if os.Getenv("GO_ENV") == "production" {
		baseURL = os.Getenv("URL_PROD")
	}

	payloadEmail := map[string]interface{}{
		"name": checkUser.Name,
		"link": fmt.Sprintf("%s/verify/%s/%s?token=%s", baseURL, role, checkUser.UUID, generateToken),
	}

	sendEmailHelper := helper.NewSendEmailHelper(s.emailLoggerRepo, s.db)
	_, err = sendEmailHelper.SendEmail(ctx, input.Email, "Email verification", "verified-user", payloadEmail, checkUser.ID)
	if err != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [send_email]")
	}

	return CheckUserResponse{
		Message: "user_found",
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, input LoginRequest, role string) (LoginResponse, error) {

	filterFindOneBy := user.User{
		Email: input.Email,
		Role:  role,
	}

	// check email
	checkUser, err := s.userRepo.FindOneBy(ctx, filterFindOneBy)
	if err != nil {
		return LoginResponse{}, errors.New("email atau password salah")
	}

	// checkLogger
	checkUserLoggerFalse, err := s.userLoggerRepo.FindAllByTheDay(uint(checkUser.ID), "login", false, input.DeviceID)
	if err != nil {
		return LoginResponse{}, errors.New("Login Log Failed")
	}
	if len(checkUserLoggerFalse) >= 3 {
		return LoginResponse{}, errors.New("Login has been limited")
	}

	checkPassword := helper.VerifyPassword(input.Password, checkUser.Password)
	if !checkPassword {
		action := "login"
		success := false
		deviceId := input.DeviceID
		ipAddress := input.IPAddress

		loggerPayload := user_logger.UserLogger{
			UserID:    uint(checkUser.ID),
			Action:    &action,
			Success:   &success,
			DeviceID:  &deviceId,
			IPAddress: &ipAddress,
		}

		err = s.db.Transaction(func(tx *gorm.DB) error {
			_, err := s.userLoggerRepo.Save(loggerPayload, tx)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return LoginResponse{}, errors.New("Login Log Failed")
		}

		return LoginResponse{}, errors.New("email atau password salah")
	}

	generateToken, err := helper.GenerateToken(checkUser.UUID, checkUser.Email, role)
	if err != nil {
		return LoginResponse{}, errors.New("Generate Token Failed")
	}

	actionFilter := "login"
	successFilter := true

	logSuccess, err := s.userLoggerRepo.FindOneBy(ctx, user_logger.UserLogger{
		UserID:  uint(checkUser.ID),
		Action:  &actionFilter,
		Success: &successFilter,
	})
	if err != nil {
		logSuccess.ID = 0
	}

	nowUnix := time.Now().Unix()

	err = s.db.Transaction(func(tx *gorm.DB) error {
		err := s.userLoggerRepo.DeleteOtherDeviceLog(uint(checkUser.ID), input.DeviceID, tx)
		if err != nil {
			return err
		}

		if len(checkUserLoggerFalse) > 0 {
			err := s.userLoggerRepo.DeleteUserLog(uint(checkUser.ID), "login", false, input.DeviceID, tx)
			if err != nil {
				return err
			}
		}

		if logSuccess.ID == 0 {

			action := "login"
			success := true
			deviceId := input.DeviceID
			ipAddress := input.IPAddress

			loggerPayload := user_logger.UserLogger{
				UserID:    uint(checkUser.ID),
				Action:    &action,
				Success:   &success,
				DeviceID:  &deviceId,
				IPAddress: &ipAddress,
			}
			_, err := s.userLoggerRepo.Save(loggerPayload, tx)
			if err != nil {
				return err
			}
		} else {

			err := s.userLoggerRepo.Update(logSuccess.ID, uint(checkUser.ID), "login", true, input.IPAddress, input.DeviceID, nowUnix, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return LoginResponse{}, err
	}

	response := LoginResponse{
		AccessToken: generateToken,
	}

	return response, nil
}

func (s *authServiceImpl) Register(ctx context.Context, input RegisterRequest, role string) (RegisterResponse, error) {
	checkUserByEmail, errEmail := s.userRepo.FindOneBy(ctx, user.User{Email: input.Email, Role: role})
	if errEmail == nil {
		if checkUserByEmail.EmailVerifiedAt != nil {
			return RegisterResponse{}, errors.New("register_failed [email_already_registered]")
		}

		resendEmailHelper := helper.NewResendEmailHelper(s.emailLoggerRepo, s.db)
		resultResendEmail, err := resendEmailHelper.ReSendEmail(ctx, checkUserByEmail, input.Email, role)
		if err != nil || !resultResendEmail.Status {
			return RegisterResponse{}, errors.New("register_failed [resend_email]")
		}

		return RegisterResponse{Message: "Resend verification email success"}, nil
	}

	checkUserByName, errName := s.userRepo.FindOneBy(ctx, user.User{Name: input.Name})
	if errName == nil && checkUserByName.EmailVerifiedAt != nil {
		return RegisterResponse{}, errors.New("register_failed [name_already_registered]")
	}

	hashPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		return RegisterResponse{}, errors.New("register_failed [hash_password]")
	}

	payloadNewUser := user.User{
		UUID:     uuid.New().String(),
		Name:     input.Name,
		Email:    input.Email,
		Password: hashPassword,
		Role:     role,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		newUser, err := s.userRepo.Save(payloadNewUser, tx)
		if err != nil {
			return fmt.Errorf("register_failed [create_user]: %w", err)
		}

		token, err := helper.GenerateToken(payloadNewUser.UUID, input.Email, role)
		if err != nil {
			return errors.New("register_failed [generate_token]")
		}

		baseURL := os.Getenv("URL_DEV")
		if os.Getenv("GO_ENV") == "production" {
			baseURL = os.Getenv("URL_PROD")
		}

		payloadEmail := map[string]interface{}{
			"name": payloadNewUser.Name,
			"link": fmt.Sprintf("%s/verify/%s/%s?token=%s", baseURL, role, payloadNewUser.UUID, token),
		}

		sendEmailHelper := helper.NewSendEmailHelper(s.emailLoggerRepo, tx)
		_, err = sendEmailHelper.SendEmail(ctx, input.Email, "Email verification", "verified-user", payloadEmail, newUser.ID)
		if err != nil {
			return fmt.Errorf("register_failed [send_email_new]: %w", err)
		}

		return nil
	})

	if err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		Message: "Register Success",
	}, nil
}
