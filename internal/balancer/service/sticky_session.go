package service

import (
	"fmt"
	"sync"
)

type stickySessionServerPool struct {
	backends      []Backend
	currentIndex  int
	mux           sync.RWMutex
	sessionCookie string
}

// GetServerPoolSize возвращает список сохраненных серверов
func (s *stickySessionServerPool) GetBackends() []Backend {
	return s.backends
}

// GetNextValidPeer рассчитывает следующий доступный сервер
func (s *stickySessionServerPool) GetNextValidPeer() Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	startIndex := s.currentIndex

	for {
		backend := s.backends[s.currentIndex]
		s.currentIndex = (s.currentIndex + 1) % len(s.backends)

		if s.sessionCookie == "" || s.sessionCookie == backend.GetURL().String() {
			return backend
		}

		if s.currentIndex == startIndex {
			break
		}
	}

	return nil
}

// AddBackend сохраняет сервер в список серверов
func (s *stickySessionServerPool) AddBackend(b Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.sessionCookie = fmt.Sprintf("session-%d", s.currentIndex%3+1)
	s.backends = append(s.backends, b)
}

// GetServerPoolSize возвращает количество сохраненных серверов
func (s *stickySessionServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
