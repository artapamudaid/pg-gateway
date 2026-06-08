package services

import (
	"payment-gateway/internal/models"
	"payment-gateway/internal/repositories"
)

type DestinationService struct {
	destRepo *repositories.DestinationRepository
}

func NewDestinationService() *DestinationService {
	return &DestinationService{
		destRepo: repositories.NewDestinationRepository(),
	}
}

func (s *DestinationService) GetAll() ([]models.Destination, error) {
	return s.destRepo.FindAll()
}

func (s *DestinationService) GetByID(id string) (models.Destination, error) {
	return s.destRepo.FindByID(id)
}

func (s *DestinationService) GetByRoutingCode(routingCode string) (models.Destination, error) {
	return s.destRepo.FindByRoutingCode(routingCode)
}

func (s *DestinationService) Create(dest *models.Destination) error {
	return s.destRepo.Create(dest)
}

func (s *DestinationService) Update(id string, inputData *models.Destination) (models.Destination, error) {
	dest, err := s.destRepo.FindByID(id)
	if err != nil {
		return dest, err
	}

	// Update fields
	dest.AppName = inputData.AppName
	dest.RoutingCode = inputData.RoutingCode
	dest.TargetURL = inputData.TargetURL
	dest.SecretToken = inputData.SecretToken
	dest.ProviderToken = inputData.ProviderToken

	err = s.destRepo.Update(&dest)
	return dest, err
}

func (s *DestinationService) Delete(id string) error {
	dest, err := s.destRepo.FindByID(id)
	if err != nil {
		return err
	}
	return s.destRepo.Delete(&dest)
}
