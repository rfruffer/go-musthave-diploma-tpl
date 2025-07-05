package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/config"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/handlers"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository/postgresql"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/services"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/router"
)

func main() {
	cfg := config.ParseFlags()

	var repo repository.StoreRepositoryInterface
	var service *services.Service
	var Handler *handlers.Handler

	db, err := postgresql.InitDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer postgresql.CloseDB(db)
	repo = postgresql.NewDBStore(db)

	service = services.NewURLService(repo)
	Handler = handlers.NewURLHandler(service, cfg.ResultHost)

	// doneCh := make(chan struct{})
	// queue1 := make(chan async.DeleteTask)

	// merged := async.FanIn(doneCh, queue1)
	// async.StartDeleteWorker(doneCh, repo, merged)
	// shortURLHandler.DeleteChan = queue1

	r := router.SetupRouter(router.Router{
		Handler:   Handler,
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
