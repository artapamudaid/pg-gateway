package config

import (
	"log"
	"os"

	"payment-gateway/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database: \n", err)
	}

	log.Println("Database terhubung!")
	DB = db

	// Konfigurasi Connection Pooling
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
	}

	// Auto-Migrate tabel-tabel yang dibutuhkan
	err = db.AutoMigrate(
		&models.Destination{},
		&models.Admin{},
		&models.CallbackLog{},
	)
	if err != nil {
		log.Fatal("Gagal melakukan migrasi database: \n", err)
	}

	// Buat Admin default jika belum ada
	models.SeedAdmin(DB)
	
	// Buat Data Dummy Destination jika belum ada
	models.SeedDestination(DB)
}
