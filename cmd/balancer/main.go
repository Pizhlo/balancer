package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Pizhlo/balancer/config"
	"github.com/Pizhlo/balancer/internal/balancer/service"
	"github.com/Pizhlo/balancer/internal/balancer/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// create service
	srv := service.New(db)

	configs, err := srv.Balancer.GetAddress(serverCtx)
	if err != nil {
		log.Fatal("unable to load configs from db: ", err)
	}

	fmt.Printf("loaded configs: %+v\n", configs)
}
