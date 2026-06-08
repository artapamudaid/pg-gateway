package repositories

import (
	"payment-gateway/config"
	"payment-gateway/internal/models"
)

type AdminRepository struct{}

func NewAdminRepository() *AdminRepository {
	return &AdminRepository{}
}

func (r *AdminRepository) FindByUsername(username string) (models.Admin, error) {
	var admin models.Admin
	err := config.DB.Where("username = ?", username).First(&admin).Error
	return admin, err
}
