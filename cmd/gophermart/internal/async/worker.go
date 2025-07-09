package async

import (
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
)

func StartOrderWorker(orderQueue <-chan string, svc *services.Service) {
	go func() {
		for orderNumber := range orderQueue {
			svc.ProcessAccrual(orderNumber)
		}
	}()
}
