package service

import (
	"sync"
)

type lcServerPool struct {
	backends []Backend
	mux      sync.RWMutex
}

// GetNextValidPeer рассчитывает следующий доступный сервер
func (s *lcServerPool) GetNextValidPeer() Backend {
	var leastConnectedPeer Backend
	for _, b := range s.backends {
		if b.IsAlive() {
			leastConnectedPeer = b
			break
		}
	}

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		if leastConnectedPeer.GetActiveConnections() > b.GetActiveConnections() {
			leastConnectedPeer = b
		}
	}
	return leastConnectedPeer
}

// AddBackend сохраняет сервер в список серверов
func (s *lcServerPool) AddBackend(b Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.backends = append(s.backends, b)
}

// GetServerPoolSize возвращает количество сохраненных серверов
func (s *lcServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

// GetServerPoolSize возвращает список сохраненных серверов
func (s *lcServerPool) GetBackends() []Backend {
	return s.backends
}
