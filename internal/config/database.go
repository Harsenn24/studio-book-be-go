package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() *gorm.DB {
	// Memuat file .env dari root project
	err := godotenv.Load()
	if err != nil {
		// Kita berikan log warning saja, karena di lingkungan production .env terkadang disuntikkan langsung via OS env
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan Environment Variable dari sistem")
	}

	// Membaca konfigurasi dari file .env dengan nilai fallback (default) jika kosong
	username := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	database := getEnv("DB_NAME", "sewa_studio")

	// Format DSN (Data Source Name) untuk GORM MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, database)

	// Membuka koneksi ke MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log query SQL otomatis muncul di terminal
	})

	if err != nil {
		log.Fatalf("Gagal terkoneksi ke database MySQL: %v", err)
	}

	// Pengaturan Connection Pool
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour * 1)
	}

	log.Println("Berhasil terkoneksi ke database MySQL menggunakan konfigurasi .env!")
	return db
}

// getEnv adalah fungsi pembantu (helper) untuk membaca env atau mengembalikan nilai default jika kosong
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
