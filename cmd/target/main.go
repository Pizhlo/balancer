package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/Pizhlo/balancer/config/target"
	"github.com/Pizhlo/balancer/internal/target/handler"
	"github.com/Pizhlo/balancer/internal/target/service"
	"github.com/Pizhlo/balancer/internal/target/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	strategy, handler, service, db := setup(ctx)

	address, err := service.Targeter.GetAddress(ctx)
	if err != nil {
		log.Fatal("unable to load configs from db: ", err)
	}

	service.CreateLogger(address, strategy)

	server := &http.Server{Addr: address, Handler: router(handler)}

	go func() {
		<-ctx.Done()
		log.Println("shutting down target")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = service.Targeter.UpdateStatus(shutdownCtx, false, address)
		if err != nil {
			log.Fatal("err while updating status: ", err)
		}

		log.Println("closing db")
		db.Close()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal("err while shutting down target: ", err)
		}
	}()

	err = service.Targeter.UpdateStatus(ctx, true, address)
	if err != nil {
		log.Fatal("err while updating status: ", err)
	}

	log.Println("starting target at", address)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		dbErr := service.Targeter.UpdateStatus(ctx, false, address)
		if dbErr != nil {
			log.Fatal("err while updating status: ", dbErr)
		}
		log.Fatal("err while starting target: ", err)
	}

}

func setup(ctx context.Context) (string, *handler.Handler, *service.Service, *postgres.TargetStore) {
	// loading config
	conf, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("unable to load config: ", err)
	}

	// creating new connection to db
	conn, err := pgxpool.New(ctx, conf.DBAddress)
	if err != nil {
		log.Fatal("unable to create connection: ", err)
	}

	// creating db
	db := postgres.New(conn)

	service := service.New(db, conf.TickerDuration, conf.SleepDuration)

	handler := handler.New(service)

	return conf.Strategy, handler, service, db
}

func router(handler *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handler.GetRequest(w, r)
	})

	r.Get("/counter", func(w http.ResponseWriter, r *http.Request) {
		handler.GetCounter(w, r)
	})

	return r
}
