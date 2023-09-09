package service

import (
	"log"
	"sync"
)

type roundRobinServerPool struct {
	backends []Backend
	mux      sync.RWMutex
	current  int
}

func (s *roundRobinServerPool) Rotate() Backend {
	log.Println("round robin: rotate")
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *roundRobinServerPool) GetNextValidPeer() Backend {
	log.Println("round robin: GetNextValidPeer")
	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []Backend {
	log.Println("round robin: GetBackends")
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b Backend) {
	log.Println("round robin: AddBackend")
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	log.Println("round robin: GetServerPoolSize")
	return len(s.backends)
}
