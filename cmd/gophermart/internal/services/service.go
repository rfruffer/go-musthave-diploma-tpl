package services

import (
	"context"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository"
)

type Service struct {
	repo repository.StoreRepositoryInterface
}

func NewURLService(repo repository.StoreRepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(login string, password string) (*models.User, error) {
	user, err := s.repo.CreateUser(context.Background(), login, password)
	if err != nil {
		return nil, err
	}
	return user, nil

}

func (s *Service) GetUserByLogin(login string) (*models.User, error) {
	user, err := s.repo.GetUserByLogin(context.Background(), login)
	if err != nil {
		return nil, err
	}
	return user, nil
}
