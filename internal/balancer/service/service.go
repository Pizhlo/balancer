package service

import (
	"context"
	"sync"

	model "github.com/Pizhlo/balancer/model/balancer"
	"github.com/pkg/errors"
)

type Balancer interface {
	GetTargets(ctx context.Context) ([]model.Target, error)
}

type Service struct {
	Balancer Balancer
	Targets  []model.Target
	mutex    sync.RWMutex
}

func New(b Balancer) *Service {
	s := &Service{Balancer: b}

	return s
}

// LoadTargets загружает из базы все доступные серверы
func (s *Service) LoadTargets(ctx context.Context) error {
	targets, err := s.Balancer.GetTargets(ctx)
	if err != nil {
		return errors.Wrap(err, "err while creating service:")
	}

	s.Targets = targets

	return nil
}
