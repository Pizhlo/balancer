package service

import (
	"net/http"
)

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type loadBalancer struct {
	serverPool ServerPool
}

const (
	RETRY_ATTEMPTED int = 0
)

func AllowRetry(r *http.Request) bool {
	if _, ok := r.Context().Value(RETRY_ATTEMPTED).(bool); ok {
		return false
	}
	return true
}

type ServerPool interface {
	GetBackends() []Backend
	GetNextValidPeer() Backend
	AddBackend(Backend)
	GetServerPoolSize() int
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextValidPeer()
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool ServerPool) LoadBalancer {
	return &loadBalancer{
		serverPool: serverPool,
	}
}
