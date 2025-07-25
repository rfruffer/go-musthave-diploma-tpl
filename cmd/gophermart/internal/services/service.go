package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/repository"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/repository/customerrors"
)

type Service struct {
	repo       repository.StoreRepositoryInterface
	orderQueue chan string
	accrualURL string
}

func NewService(repo repository.StoreRepositoryInterface, accrualURL string, orderQueue chan string) *Service {
	return &Service{
		repo:       repo,
		accrualURL: accrualURL,
		orderQueue: orderQueue,
	}
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

func (s *Service) SaveNewOrder(ctx context.Context, userID, orderNumber string) (int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, err
	}
	err = s.repo.InsertOrder(ctx, uid, orderNumber)
	if err != nil {
		if errors.Is(err, customerrors.ErrOrderAlreadyUploadedBySameUser) {
			return 200, err
		}
		if errors.Is(err, customerrors.ErrOrderUploadedByAnotherUser) {
			return 409, err
		}
		return 500, err
	}
	log.Printf("Status: %v", orderNumber)
	s.EnqueueOrderForProcessing(orderNumber)
	return 202, nil
}

func (s *Service) EnqueueOrderForProcessing(orderNumber string) {
	select {
	case s.orderQueue <- orderNumber:
	default:
		log.Printf("order queue full, dropping order: %s", orderNumber)
	}
}

func (s *Service) ProcessAccrual(orderNumber string) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualURL, orderNumber)

	for {
		resp, err := http.Get(url)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retry := time.Second * 5
			if val := resp.Header.Get("Retry-After"); val != "" {
				if sec, err := strconv.Atoi(val); err == nil {
					retry = time.Duration(sec) * time.Second
				}
			}
			resp.Body.Close()
			time.Sleep(retry)
			continue
		}

		if resp.StatusCode == http.StatusNoContent {
			resp.Body.Close()
			return
		}

		var res models.AccrualResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if res.Status == "REGISTERED" || res.Status == "PROCESSING" {
			time.Sleep(3 * time.Second)
			continue
		}

		if res.Status == "PROCESSED" {
			log.Printf("accrual processed: %s +%.2f", res.Order, *res.Accrual)
			if err := s.repo.UpdateOrderAccrual(context.Background(), res.Order, res.Status, *res.Accrual); err != nil {
				log.Printf("failed to update accrual: %v", err)
			}
			return
		}
	}
}

func (s *Service) GetUserOrders(ctx context.Context, userID string) ([]models.Order, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	orders, err := s.repo.GetOrdersByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *Service) Withdraw(ctx context.Context, userID, order string, amount float64) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	if !IsValidLuhn(order) {
		return customerrors.ErrInvalidOrderNumber
	}
	return s.repo.Withdraw(ctx, uid, order, amount)
}

func (s *Service) GetWithdrawals(ctx context.Context, userID string) ([]models.Withdrawal, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetWithdrawals(ctx, uid)
}

func (s *Service) GetUserBalance(ctx context.Context, userID string) (*models.Balance, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserBalance(ctx, uid)
}
