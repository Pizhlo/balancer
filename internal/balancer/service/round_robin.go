package service

import (
	"sync"
)

type roundRobinServerPool struct {
	backends []Backend
	mux      sync.RWMutex
	current  int
}

func (s *roundRobinServerPool) Rotate() Backend {
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *roundRobinServerPool) GetNextValidPeer() Backend {
	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []Backend {
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b Backend) {
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
