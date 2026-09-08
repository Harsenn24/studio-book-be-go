package UserRepository

import (
	"errors"
	"time"

	"context"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type UserRepository interface {
	Save(u User, tx *gorm.DB) (User, error)
	FindByEmail(email string) (User, error)
	FindOneBy(ctx context.Context, filter User) (User, error)
}

type userRepositoryImpl struct {
	db *gorm.DB // <--- SEKARANG MENGGUNAKAN KONEKSI DATABASE ASLI GORM, BUKAN SLICE MEMORI LAGI
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

func (r *userRepositoryImpl) Save(u User, tx *gorm.DB) (User, error) {
	// Isi timestamp secara otomatis sebelum disimpan ke database
	now := time.Now().Unix()
	u.CreatedAt = now
	u.UpdatedAt = now

	// GORM otomatis menjalankan: INSERT INTO users (...) VALUES (...);
	err := r.db.Create(&u).Error
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (r *userRepositoryImpl) FindOneBy(ctx context.Context, filter User) (User, error) {
	var user User

	baseQuery := "SELECT * FROM users WHERE "
	var conditions []string
	var args []interface{}

	v := reflect.ValueOf(filter)
	t := reflect.TypeOf(filter)

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(fieldType.Name)
		}

		if !fieldVal.IsZero() {
			conditions = append(conditions, fmt.Sprintf("%s = ?", dbTag))
			args = append(args, fieldVal.Interface())
		}
	}

	if len(conditions) == 0 {
		return User{}, errors.New("kriteria pencarian tidak boleh kosong")
	}

	finalQuery := baseQuery + strings.Join(conditions, " AND ") + " LIMIT 1"

	err := r.db.WithContext(ctx).Raw(finalQuery, args...).Scan(&user).Error
	if err != nil {
		return User{}, err
	}
	if user.ID == 0 {
		return User{}, errors.New("user tidak ditemukan")
	}

	return user, nil
}

func (r *userRepositoryImpl) FindByEmail(email string) (User, error) {
	var user User

	// GORM otomatis menjalankan: SELECT * FROM users WHERE email = ? LIMIT 1;
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		// Jika data tidak ditemukan, GORM mengembalikan error gorm.ErrRecordNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, errors.New("user dengan email tersebut tidak ditemukan")
		}
		return User{}, err // Return error database lainnya (misal: koneksi putus)
	}

	return user, nil
}
