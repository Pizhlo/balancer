package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	handler, service, db := setup()

	address, err := service.Targeter.GetAddress(context.TODO())
	if err != nil {
		log.Fatal("unable to load configs from db: ", err)
	}

	server := &http.Server{Addr: address, Handler: router(handler)}

	log.Println("starting server at", address)

	err = service.Targeter.UpdateStatus(context.TODO(), true, address)
	if err != nil {
		log.Fatal("err while updating status: ", err)
	}

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		err = service.Targeter.UpdateStatus(context.TODO(), false, address)
		if err != nil {
			log.Fatal("err while updating status: ", err)
		}
	}

	//wait for termination signal and register database & http server clean-up operations
	wait := gracefulShutdown(context.TODO(), 2*time.Second, map[string]operation{

		"http-server": func(ctx context.Context) error {
			err = service.Targeter.UpdateStatus(context.TODO(), false, address)
			return server.Shutdown(context.Background())
		},
		"db": func(ctx context.Context) error {
			db.Close()
			return nil
		},
	})

	<-wait

}

func setup() (*handler.Handler, *service.Service, *postgres.TargetStore) {
	// loading config
	conf, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("unable to load config: ", err)
	}

	// creating new connection to db
	conn, err := pgxpool.New(context.TODO(), conf.DBAddress)
	if err != nil {
		log.Fatal("unable to create connection: ", err)
	}

	// creating db
	db := postgres.New(conn)

	service := service.New(db, conf.TickerDuration)
	service.SleepDuration = conf.SleepDuration

	handler := handler.New(service)

	return handler, service, db
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

// operation is a clean up function on shutting down
type operation func(ctx context.Context) error

// gracefulShutdown waits for termination syscalls and doing clean up operations after received it
func gracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		// add any other syscalls that you want to be notified with
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("shutting down")

		// set timeout for the ops to be done to prevent system hang
		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Printf("timeout %d ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		// Do the operations asynchronously to save time
		for key, op := range ops {
			wg.Add(1)
			innerOp := op
			innerKey := key
			go func() {
				defer wg.Done()

				log.Printf("cleaning up: %s", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Printf("%s: clean up failed: %s", innerKey, err.Error())
					return
				}

				log.Printf("%s was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()

		close(wait)
	}()

	return wait
}
