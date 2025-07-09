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

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository/customErrors"
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
	err := s.repo.InsertOrder(ctx, userID, orderNumber)
	if err != nil {
		if errors.Is(err, customErrors.ErrOrderAlreadyUploadedBySameUser) {
			return 200, err
		}
		if errors.Is(err, customErrors.ErrOrderUploadedByAnotherUser) {
			return 409, err
		}
		return 500, err
	}
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

		var res models.Order
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
		log.Printf("accrual status: order=%s status=%s accrual=%.2f", res.Number, res.Status, *res.Accrual)

		if err := s.repo.UpdateOrderAccrual(context.Background(), res.Number, res.Status, *res.Accrual); err != nil {
			log.Printf("failed to update accrual: %v", err)
		}
		return
	}
}

func (s *Service) GetUserOrders(ctx context.Context, userID string) ([]models.Order, error) {
	orders, err := s.repo.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
