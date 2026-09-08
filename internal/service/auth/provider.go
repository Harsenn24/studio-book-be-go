package auth

import (
	emailLoggerRepo "studio-book-be-go/internal/repository/email_logger"
	user "studio-book-be-go/internal/repository/user"
	user_logger "studio-book-be-go/internal/repository/user_logger"

	"gorm.io/gorm"
)

type AuthModule struct {
	Handler *AuthHandler
}

func InitAuthModule(db *gorm.DB) *AuthModule {
	userRepo := user.NewUserRepository(db)
	userLoggerRepo := user_logger.NewUserLoggerRepository(db)
	emailLoggerRepo := emailLoggerRepo.NewEmailLoggerRepository(db)

	service := NewAuthService(
		db,
		userRepo,
		userLoggerRepo,
		emailLoggerRepo,
	)
	handler := NewAuthHandler(service)

	return &AuthModule{
		Handler: handler,
	}
}
