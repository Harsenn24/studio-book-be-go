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
	CheckUser(ctx context.Context, input CheckUserRequest) (CheckUserResponse, error)
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

func (s *authServiceImpl) CheckUser(ctx context.Context, input CheckUserRequest) (CheckUserResponse, error) {
	filterFindOneBy := user.User{
		Email: input.Email,
		Role:  input.Role,
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

	generateToken, err := helper.GenerateToken(checkUser.UUID, checkUser.Email, input.Role)
	if err != nil {
		return CheckUserResponse{}, errors.New("check_user_failed [generate_token]")
	}

	baseURL := os.Getenv("URL_DEV")
	if os.Getenv("GO_ENV") == "production" {
		baseURL = os.Getenv("URL_PROD")
	}

	payloadEmail := map[string]interface{}{
		"name": checkUser.Name,
		"link": fmt.Sprintf("%s/verify/%s/%s?token=%s", baseURL, input.Role, checkUser.UUID, generateToken),
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
	filterFindOneByEmailAndRole := user.User{
		Email: input.Email,
		Role:  role,
	}

	// check email
	checkUserByEmail, err := s.userRepo.FindOneBy(ctx, filterFindOneByEmailAndRole)
	if err != nil {

		filterFindOneByName := user.User{
			Name: input.Name,
		}

		hashPassword, err := helper.HashPassword(input.Password)
		if err != nil {
			return RegisterResponse{}, errors.New("hash password failed")
		}

		payloadNewUser := user.User{
			UUID:     uuid.New().String(),
			Name:     input.Name,
			Email:    input.Email,
			Password: hashPassword,
			Role:     role,
		}

		checkUserByName, errByName := s.userRepo.FindOneBy(ctx, filterFindOneByName)
		if errByName == nil && checkUserByName.EmailVerifiedAt != nil {
			return RegisterResponse{}, errors.New("name already registered")
		}

		if errByName != nil {

			// transaction cuma untuk operasi DB
			err = s.db.Transaction(func(tx *gorm.DB) error {
				_, err := s.userRepo.Save(payloadNewUser, tx)
				return err
			})

			if err != nil {
				return RegisterResponse{}, fmt.Errorf("failed to create user: %w", err)
			}

			// generate token & kirim email di luar transaction
			token, err := helper.GenerateToken(payloadNewUser.UUID, input.Email, role)
			if err != nil {
				return RegisterResponse{}, errors.New("generate token failed")
			}

			baseURL := os.Getenv("URL_DEV")
			if os.Getenv("GO_ENV") == "production" {
				baseURL = os.Getenv("URL_PROD")
			}

			payloadEmail := map[string]interface{}{
				"name": payloadNewUser.Name,
				"link": fmt.Sprintf("%s/verify/%s/%s?token=%s", baseURL, role, payloadNewUser.UUID, token),
			}

			sendEmailHelper := helper.NewSendEmailHelper(s.emailLoggerRepo, s.db)
			_, err = sendEmailHelper.SendEmail(ctx, input.Email, "Email verification", "verified-user", payloadEmail, payloadNewUser.ID)
			if err != nil {
				return RegisterResponse{}, errors.New("send email failed")
			}

			return RegisterResponse{
				Message: "Register Success",
			}, nil
		}

		err = s.db.Transaction(func(tx *gorm.DB) error {
			_, err := s.userRepo.Save(payloadNewUser, tx)
			return err
		})

		if err != nil {
			return RegisterResponse{}, fmt.Errorf("failed to create user: %w", err)
		}

	}

	if checkUserByEmail.EmailVerifiedAt != nil {
		return RegisterResponse{}, errors.New("Email already registered")
	}

	resendEmailHelper := helper.NewResendEmailHelper(s.emailLoggerRepo, s.db)
	resultResendEmail, err := resendEmailHelper.ReSendEmail(ctx, checkUserByEmail, input.Email, role)
	if err != nil {
		return RegisterResponse{}, errors.New("Send Email Failed")
	}

	if !resultResendEmail.Status {
		return RegisterResponse{}, errors.New("Send Email Failed")
	}

	return RegisterResponse{}, nil
}
