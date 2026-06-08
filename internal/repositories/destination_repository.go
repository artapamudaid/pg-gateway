package repositories

import (
	"payment-gateway/config"
	"payment-gateway/internal/models"
)

type DestinationRepository struct{}

func NewDestinationRepository() *DestinationRepository {
	return &DestinationRepository{}
}

func (r *DestinationRepository) FindAll() ([]models.Destination, error) {
	var dests []models.Destination
	err := config.DB.Find(&dests).Error
	return dests, err
}

func (r *DestinationRepository) FindByID(id string) (models.Destination, error) {
	var dest models.Destination
	err := config.DB.First(&dest, id).Error
	return dest, err
}

func (r *DestinationRepository) FindByRoutingCode(routingCode string) (models.Destination, error) {
	var dest models.Destination
	err := config.DB.Where("routing_code = ?", routingCode).First(&dest).Error
	return dest, err
}

func (r *DestinationRepository) Create(dest *models.Destination) error {
	return config.DB.Create(dest).Error
}

func (r *DestinationRepository) Update(dest *models.Destination) error {
	return config.DB.Save(dest).Error
}

func (r *DestinationRepository) Delete(dest *models.Destination) error {
	return config.DB.Delete(dest).Error
}
