package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pizhlo/balancer/config"
	"github.com/Pizhlo/balancer/internal/balancer/handler"
	"github.com/Pizhlo/balancer/internal/balancer/service"
	"github.com/Pizhlo/balancer/internal/balancer/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// loading config
	conf, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("unable to load config: ", err)
	}

	// creating new connection to db
	conn, err := pgxpool.New(serverCtx, conf.DBAddress)
	if err != nil {
		log.Fatal("unable to create connection: ", err)
	}

	// creating db
	db := postgres.New(conn)
	defer db.Close()

	service := service.New(db)

	handler := handler.New(service)

	addr := fmt.Sprintf("localhost:%s", conf.ServerPort)
	server := &http.Server{Addr: addr, Handler: router(handler)}

	log.Println("starting balancer at", addr)

	configs, err := service.Balancer.GetConfig(serverCtx)
	if err != nil {
		log.Fatal("unable to load configs from db: ", err)
	}

	service.Configs = configs
	log.Printf("loaded configs: %+v\n", service.Configs)

	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal("error while shutdown server: ", err)
		}
		serverStopCtx()
	}()

	// starting server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("error while starting server: ", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func router(handler *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handler.GetRequest(w, r)
	})

	return r
}
