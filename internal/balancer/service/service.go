package service

import (
	"context"
	"log"
	"sync"
	"time"

	model "github.com/Pizhlo/balancer/model/balancer"
)

type Balancer interface {
	GetConfig(ctx context.Context) ([]model.ConfigDB, error)
}

type Service struct {
	Balancer Balancer
	Configs  []model.ConfigDB
	Counter  int
	mutex    sync.Mutex
}

func New(b Balancer) *Service {
	return &Service{Balancer: b}
}

func (s *Service) Handle() {
	s.increment()
	time.Sleep(1 * time.Second)
	s.decrement()
}

func (s *Service) increment() {
	s.mutex.Lock()
	s.Counter++
	s.mutex.Unlock()
}

func (s *Service) decrement() {
	s.mutex.Lock()
	s.Counter--
	s.mutex.Unlock()
}

func (s *Service) Print(ticker *time.Ticker, done chan bool) {
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.log()
		}
	}
}

func (s *Service) log() {
	log.Println("current number of requests:", s.Counter)
}
