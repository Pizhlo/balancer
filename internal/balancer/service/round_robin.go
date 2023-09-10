package service

import (
	"sync"
)

type roundRobinServerPool struct {
	backends []Backend
	mux      sync.RWMutex
	current  int
}

// Rotate увеличивает текущее значение и возвращает следующий сервер в строке
func (s *roundRobinServerPool) Rotate() Backend {
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

// GetNextValidPeer рассчитывает следующий доступный сервер
func (s *roundRobinServerPool) GetNextValidPeer() Backend {
	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

// GetServerPoolSize возвращает список сохраненных серверов
func (s *roundRobinServerPool) GetBackends() []Backend {
	return s.backends
}

// AddBackend сохраняет сервер в список серверов
func (s *roundRobinServerPool) AddBackend(b Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.backends = append(s.backends, b)
}

// GetServerPoolSize возвращает количество сохраненных серверов
func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
