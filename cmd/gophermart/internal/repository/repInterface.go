package repository

import (
	"context"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
)

type StoreRepositoryInterface interface {
	CreateUser(ctx context.Context, login, password string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}
