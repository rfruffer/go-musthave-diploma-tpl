package async

import (
	"log"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
)

func StartOrderWorker(orderQueue <-chan string, svc *services.Service) {
	go func() {
		log.Println("⚙️ order worker started")
		for orderNumber := range orderQueue {
			log.Printf("📦 processing order from queue: %s", orderNumber)
			svc.ProcessAccrual(orderNumber)
		}
	}()
}
