package service

import (
	"context"
	"sync"

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

func (s *Service) Increment() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Counter++
	defer s.decrement()
}

func (s *Service) decrement() {
	s.Counter--
}
