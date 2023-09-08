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
	Balancer      Balancer
	Configs       []model.ConfigDB
	Counter       int
	mutex         sync.Mutex
	SleepDuration time.Duration
}

func New(b Balancer, tickerDuration time.Duration) *Service {
	s := &Service{Balancer: b}

	done := make(chan bool)
	s.startTicker(tickerDuration, done)

	return s
}

func (s *Service) Handle() {
	s.increment()
	time.Sleep(s.SleepDuration)
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

func (s *Service) log() {
	log.Println("current number of requests:", s.Counter)
}

func (s *Service) startTicker(d time.Duration, done chan bool) *time.Ticker {
	tick := time.NewTicker(d)

	for {
		select {
		case <-tick.C:
			s.log()
		case <-done:
			break
		default:
		}
	}
}
