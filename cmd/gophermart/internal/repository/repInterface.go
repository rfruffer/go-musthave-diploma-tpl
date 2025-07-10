package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
)

type StoreRepositoryInterface interface {
	// Аутентификация
	CreateUser(ctx context.Context, login, password string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	// Работа с заказами
	InsertOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error
	UpdateOrderAccrual(ctx context.Context, orderNumber, status string, accrual float64) error
	GetPendingOrders(ctx context.Context) ([]string, error)
	GetOrdersByUser(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	Withdraw(ctx context.Context, userID uuid.UUID, order string, amount float64) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*models.Balance, error)
}
