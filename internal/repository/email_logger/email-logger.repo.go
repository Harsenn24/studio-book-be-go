package EmailLoggerRepository

import (
	"time"
	"gorm.io/gorm"
)

type EmailLoggerRepository interface {
	Save(u EmailLogger, tx *gorm.DB) (EmailLogger, error)
	FindAllByEmailAndType(email string, typeEmail string) ([]EmailLogger, error)
}

type emailLoggerRepositoryImpl struct {
	db *gorm.DB 
}

func NewEmailLoggerRepository(db *gorm.DB) EmailLoggerRepository {
	return &emailLoggerRepositoryImpl{
		db: db,
	}
}

func (r *emailLoggerRepositoryImpl) Save(u EmailLogger, tx *gorm.DB) (EmailLogger, error) {
	now := time.Now().Unix()
	u.CreatedAt = now
	u.UpdatedAt = now

	err := r.db.Create(&u).Error
	if err != nil {
		return EmailLogger{}, err
	}

	return u, nil
}

func (r *emailLoggerRepositoryImpl) FindAllByEmailAndType(email string, typeEmail string) ([]EmailLogger, error) {
	var emailLogs []EmailLogger
	err := r.db.Where("email = ? AND type = ?", email, typeEmail).Find(&emailLogs).Error
	if err != nil {
		return nil, err
	}

	return emailLogs, nil
}


