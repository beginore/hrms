package service

import (
	"context"
	"hrms/internal/feature/city/repository"
)

type CityResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type CityService interface {
	ListCities(ctx context.Context) ([]repository.City, error)
}

type cityService struct {
	repo repository.CityRepository
}

func NewCityService(repo repository.CityRepository) CityService {
	return &cityService{repo: repo}
}

func (s *cityService) ListCities(ctx context.Context) ([]repository.City, error) {
	return s.repo.GetAll(ctx)
}
