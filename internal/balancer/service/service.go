package service

import (
	"context"

	model "github.com/Pizhlo/balancer/model/balancer"
)

type Balancer interface {
	GetAddress(ctx context.Context) ([]model.ConfigDB, error)
}

type Service struct {
	Balancer Balancer
}

func New(b Balancer) *Service {
	return &Service{Balancer: b}
}
