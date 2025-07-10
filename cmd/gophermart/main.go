// cmd/gophermart/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/config"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/async"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/handlers"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/repository/postgresql"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/internal/services"
	"github.com/rfruffer/go-musthave-diploma-tpl/cmd/gophermart/router"
)

func main() {
	cfg := config.ParseFlags()

	db, err := postgresql.InitDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer postgresql.CloseDB(db)

	repo := postgresql.NewDBStore(db)
	orderQueue := make(chan string, 100)

	service := services.NewService(repo, cfg.Accrual, orderQueue)
	handler := handlers.NewHandler(service, cfg.SecretKey)

	// запуск воркера
	async.StartOrderWorker(orderQueue, service)

	r := router.SetupRouter(router.Router{
		Handler:   handler,
		SecretKey: cfg.SecretKey,
	})

	server := &http.Server{
		Addr:    cfg.StartHost,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on %s", cfg.StartHost)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down server...")

	if err := server.Close(); err != nil {
		log.Printf("error shutting down server: %v", err)
	}

	log.Println("server stopped gracefully")
}
