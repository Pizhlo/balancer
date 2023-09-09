package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	config "github.com/Pizhlo/balancer/config/balancer"
	"github.com/Pizhlo/balancer/internal/balancer/service"
	"github.com/Pizhlo/balancer/internal/balancer/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	conf, service, db := setup()

	err := service.LoadTargets(context.TODO())
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("unable to load targets: ", err)
	}

	log.Printf("successfully loaded targets: %+v\n", service.Targets)

	addr := fmt.Sprintf("localhost:%s", conf.BalancerPort)
	server := &http.Server{Addr: addr, Handler: router()}

	log.Println("starting balancer at", addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("error while starting server: ", err)
	}

	wait := gracefulShutdown(context.TODO(), 2*time.Second, map[string]operation{

		"http-server": func(ctx context.Context) error {
			return server.Shutdown(context.Background())
		},
		"db": func(ctx context.Context) error {
			db.Close()
			return nil
		},
	})

	<-wait
}

func setup() (config.Config, *service.Service, *postgres.BalancerStore) {
	conf, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("unable to load config: ", err)
	}

	conn, err := pgxpool.New(context.TODO(), conf.DBAddress)
	if err != nil {
		log.Fatal("unable to create connection: ", err)
	}

	db := postgres.New(conn)

	service := service.New(db)

	return conf, service, db
}

func router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		//handler.GetRequest(w, r)
	})

	return r
}

type operation func(ctx context.Context) error

func gracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("shutting down")

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Printf("timeout %d ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

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
