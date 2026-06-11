package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

type Destination struct {
	gorm.Model
	AppName       string `gorm:"size:100;not null"`
	Environment   string `gorm:"size:20;uniqueIndex:idx_routing_env;not null;default:'production'"` // sandbox / production
	RoutingCode   string `gorm:"size:50;uniqueIndex:idx_routing_env;not null"` // Contoh: APP1, SHOP
	TargetURL     string `gorm:"not null"`                     // URL sistem tujuan
	SecretToken   string `gorm:"not null"`                     // Handshake token (X-Gateway-Auth)
	ProviderToken string `gorm:"not null;default:''"`          // Token dari Flip/Midtrans untuk verifikasi webhook masuk
}

type CallbackLog struct {
	gorm.Model
	Environment string
	ReferenceID string
	RoutingCode string
	TargetURL   string
	Payload     string
	StatusCode  int
}

// Seeder untuk Admin awal
func SeedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&Admin{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Create(&Admin{
			Username: "admin",
			Password: string(hash),
		})
	}
}

// Seeder untuk Destination (Data Dummy)
func SeedDestination(db *gorm.DB) {
	var count int64
	db.Model(&Destination{}).Count(&count)
	if count == 0 {
		db.Create(&Destination{
			AppName:       "Demo App",
			Environment:   "production",
			RoutingCode:   "DEMO",
			TargetURL:     "https://webhook.site/demo-destination",
			SecretToken:   "demo-secret",
			ProviderToken: "demo-provider-token",
		})
	}
}
