package service

import (
	"context"
	"log"
	"sync"
	"time"

	model "github.com/Pizhlo/balancer/model/balancer"
)

type Targeter interface {
	GetConfig(ctx context.Context) ([]model.ConfigDB, error)
}

type Service struct {
	Targeter      Targeter
	Configs       []model.ConfigDB
	counter       int
	mutex         sync.RWMutex
	SleepDuration time.Duration
}

func New(t Targeter, tickerDuration time.Duration) *Service {
	s := &Service{Targeter: t}

	done := make(chan bool)
	go s.startTicker(tickerDuration, done)

	return s
}

func (s *Service) Handle() {
	s.increment()
	time.Sleep(s.SleepDuration)
	s.decrement()
}

func (s *Service) GetCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.counter
}

func (s *Service) increment() {
	s.mutex.Lock()
	s.counter++
	s.mutex.Unlock()
}

func (s *Service) decrement() {
	s.mutex.Lock()
	s.counter--
	s.mutex.Unlock()
}

func (s *Service) log() {
	log.Println("current number of requests:", s.counter)
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
