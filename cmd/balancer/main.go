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

	service, db, serverPool, loadBalancer, server := setup(ctx)

	err := service.LoadTargets(ctx)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("unable to load targets: ", err)
	}

	log.Printf("successfully loaded targets: %+v\n", service.Targets)

	setupBackend(service, serverPool, loadBalancer)

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

	log.Println("starting balancer at", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("error while starting server: ", err)
	}
}

func setup(ctx context.Context) (*balancer.Service, *postgres.BalancerStore, balancer.ServerPool, balancer.LoadBalancer, *http.Server) {
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

	serverPool, err := balancer.NewServerPool(conf.Strategy)
	if err != nil {
		log.Fatal("err while creating server pool: ", err)
	}

	loadBalancer := balancer.NewLoadBalancer(serverPool)

	addr := fmt.Sprintf("localhost:%s", conf.BalancerPort)
	server := &http.Server{Addr: addr, Handler: router(loadBalancer)}

	return service, db, serverPool, loadBalancer, server
}

func router(lb balancer.LoadBalancer) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		lb.Serve(w, r)
	})
	r.Get("/counter", func(w http.ResponseWriter, r *http.Request) {
		lb.Serve(w, r)
	})

	return r
}

func setupBackend(service *balancer.Service, serverPool balancer.ServerPool, loadBalancer balancer.LoadBalancer) {
	for _, target := range service.Targets {
		addr := "http://" + target.Address
		endpoint, err := url.Parse(addr)
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
}
