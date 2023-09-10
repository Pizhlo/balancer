package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	config "github.com/Pizhlo/balancer/config/balancer"
	balancer "github.com/Pizhlo/balancer/internal/balancer/service"
	"github.com/Pizhlo/balancer/internal/balancer/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf, service, db := setup(ctx)

	err := service.LoadTargets(ctx)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("unable to load targets: ", err)
	}

	log.Printf("successfully loaded targets: %+v\n", service.Targets)

	addr := fmt.Sprintf("localhost:%s", conf.BalancerPort)
	server := &http.Server{Addr: addr, Handler: router()}

	serverPool, err := balancer.NewServerPool(conf.Strategy)
	if err != nil {
		log.Fatal(err.Error())
	}

	loadBalancer := balancer.NewLoadBalancer(serverPool)

	for _, target := range service.Targets {
		endpoint, err := url.Parse(target.Address)
		if err != nil {
			log.Fatalf("err while parsing url %s: %v\n", target.Address, err)
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := balancer.NewBackend(endpoint, rp)

		serverPool.AddBackend(backendServer)

		rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Println("error handling the request; host: ", endpoint.Host, "err: ", e)
			backendServer.SetAlive(false)

			if !balancer.AllowRetry(request) {

				log.Println("Max retry attempts reached, terminating; address: ", request.RemoteAddr, "path: ", request.URL.Path)
				http.Error(writer, "Service not available", http.StatusServiceUnavailable)
				return
			}

			log.Println("Attempting retry; address: ", request.RemoteAddr, "URL: ", request.URL.Path, "retry: ", true)
			loadBalancer.Serve(
				writer,
				request.WithContext(
					context.WithValue(request.Context(), balancer.RETRY_ATTEMPTED, true),
				),
			)
		}
	}

	go balancer.LauchHealthCheck(ctx, serverPool)

	go func() {
		<-ctx.Done()
		log.Println("shutting down balancer")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("closing db")
		db.Close()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal("err while shutting down balancer: ", err)
		}
	}()

	log.Println("starting balancer at", addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("error while starting server: ", err)
	}
}

func setup(ctx context.Context) (config.Config, *balancer.Service, *postgres.BalancerStore) {
	conf, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("unable to load config: ", err)
	}

	conn, err := pgxpool.New(ctx, conf.DBAddress)
	if err != nil {
		log.Fatal("unable to create connection: ", err)
	}

	db := postgres.New(conn)

	service := balancer.New(db)

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
