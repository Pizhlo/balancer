package service

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"
)

// LauchHealthCheck запускает тикер и горутину, проверяющую соединение с сервером
func LauchHealthCheck(ctx context.Context, sp ServerPool) {
	t := time.NewTicker(time.Second * 20)
	log.Println("Starting health check...")
	for {
		select {
		case <-t.C:
			go HealthCheck(ctx, sp)
		case <-ctx.Done():
			log.Println("Closing Health Check")
			return
		}
	}
}

// HealthCheck проверяет соединение с сервером и устаналивает статус
func HealthCheck(ctx context.Context, s ServerPool) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.GetBackends() {
		b := b
		requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		defer stop()
		status := "up"
		go IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			log.Println("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}

		log.Println("URL status: url: ", b.GetURL(), "status:", status)

	}
}

// IsBackendAlive отправяет запрос на сервер и записывает в канал статус сервера
func IsBackendAlive(ctx context.Context, aliveChannel chan bool, u *url.URL) {
	log.Println("IsBackendAlive; url: ", u)
	client := http.Client{}

	resp, err := client.Get(u.String())

	if err != nil {
		log.Println("unable to make http request. err: ", err)
		aliveChannel <- false
	}

	aliveChannel <- true
	log.Println("IsBackendAlive", u.String(), " successful")
	defer resp.Body.Close()
}
