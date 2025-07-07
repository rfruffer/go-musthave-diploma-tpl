package async

import (
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
)

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func StartOrderWorker(orderQueue <-chan string, svc *services.Service) {
	go func() {
		for orderNumber := range orderQueue {
			svc.ProcessAccrual(orderNumber)
		}
	}()
}
