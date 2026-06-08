package repositories

import (
	"payment-gateway/config"
	"payment-gateway/internal/models"
)

type LogRepository struct{}

func NewLogRepository() *LogRepository {
	return &LogRepository{}
}

func (r *LogRepository) Create(log *models.CallbackLog) error {
	return config.DB.Create(log).Error
}
