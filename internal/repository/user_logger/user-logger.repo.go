package UserLoggerRepository

import (
	"errors"
	"context"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

type UserLoggerRepository interface {
	Save(u UserLogger, tx *gorm.DB) (UserLogger, error)
	FindAllByTheDay(UserId uint, Action string, Success bool, DeviceId string) ([]UserLogger, error)
	DeleteOtherDeviceLog(user_id uint, device_id string, tx *gorm.DB) error
	DeleteUserLog(user_id uint, action string, success bool, device_id string, tx *gorm.DB) error
	Update(id uint, user_id uint, action string, success bool, ip_address string, device_id string, updated_at int64, tx *gorm.DB) error
	FindOneBy(ctx context.Context, filter UserLogger) (UserLogger, error)
}

type userLoggerRepositoryImpl struct {
	db *gorm.DB
}

func NewUserLoggerRepository(db *gorm.DB) UserLoggerRepository {
	return &userLoggerRepositoryImpl{
		db: db,
	}
}

func (r *userLoggerRepositoryImpl) Save(ul UserLogger, tx *gorm.DB) (UserLogger, error) {
	now := time.Now().Unix()
	ul.CreatedAt = now
	ul.UpdatedAt = now

	db := r.db
	if tx != nil {
		db = tx 
	}

	err := db.Create(&ul).Error
	if err != nil {
		return UserLogger{}, err
	}

	return ul, nil
}

func (r *userLoggerRepositoryImpl) FindAllByTheDay(UserId uint, Action string, Success bool, DeviceId string) ([]UserLogger, error) {
	var logs []UserLogger

	now := time.Now()

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()

	err := r.db.Where(
		"user_id = ? AND action = ? AND success = ? AND device_id = ? AND created_at BETWEEN ? AND ?",
		UserId, // dari parameter ul
		Action,
		Success,
		DeviceId,
		startOfToday,
		endOfToday,
	).Find(&logs).Error

	if err != nil {
		return nil, err
	}

	return logs, nil

}

func (r *userLoggerRepositoryImpl) DeleteOtherDeviceLog(user_id uint, device_id string, tx *gorm.DB) error {
	db := r.db
	if tx != nil {
		db = tx 
	}

	err := db.Where("user_id = ? AND device_id != ?", user_id, device_id).
		Delete(&UserLogger{}).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (r *userLoggerRepositoryImpl) DeleteUserLog(user_id uint, action string, success bool, device_id string, tx *gorm.DB) error {
	db := r.db
	if tx != nil {
		db = tx 
	}

	err := db.Where("user_id = ? AND device_id = ? AND action = ? AND success = ?", user_id, device_id, action, success).
		Delete(&UserLogger{}).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (r *userLoggerRepositoryImpl) Update(id uint, user_id uint, action string, success bool, ip_address string, device_id string, updated_at int64, tx *gorm.DB) error {
	db := r.db
	if tx != nil {
		db = tx 
	}

	err := db.Model(&UserLogger{}).Where("id = ?", id).Updates(map[string]interface{}{
		"user_id":    user_id,
		"action":     action,
		"success":    success,
		"ip_address": ip_address,
		"device_id":  device_id,
		"updated_at": updated_at,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *userLoggerRepositoryImpl) FindOneBy(ctx context.Context, filter UserLogger) (UserLogger, error) {
	var logger UserLogger

	baseQuery := "SELECT * FROM user_loggers WHERE " 
	var conditions []string
	var args []interface{}

	v := reflect.ValueOf(filter)
	t := reflect.TypeOf(filter)

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			gormTag := fieldType.Tag.Get("gorm")
			if strings.Contains(gormTag, "column:") {
				part := strings.Split(gormTag, "column:")
				if len(part) > 1 {
					dbTag = strings.Split(part[1], ";")[0]
				}
			}
		}
		if dbTag == "" {
			dbTag = strings.ToLower(fieldType.Name)
		}

		if fieldVal.Kind() == reflect.Ptr {
			if !fieldVal.IsNil() { 
				conditions = append(conditions, fmt.Sprintf("%s = ?", dbTag))
				args = append(args, fieldVal.Elem().Interface()) 
			}
		} else {
			if !fieldVal.IsZero() {
				conditions = append(conditions, fmt.Sprintf("%s = ?", dbTag))
				args = append(args, fieldVal.Interface())
			}
		}
	}

	if len(conditions) == 0 {
		return UserLogger{}, errors.New("kriteria pencarian tidak boleh kosong")
	}

	finalQuery := baseQuery + strings.Join(conditions, " AND ") + " LIMIT 1"

	err := r.db.WithContext(ctx).Raw(finalQuery, args...).Scan(&logger).Error
	if err != nil {
		return UserLogger{}, err
	}

	if logger.ID == 0 {
		return UserLogger{}, errors.New("log tidak ditemukan")
	}

	return logger, nil
}
