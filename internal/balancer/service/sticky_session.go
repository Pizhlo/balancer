package service

import "sync"

type stickySessionServerPool struct {
	backends []Backend
	mux      sync.RWMutex
}

// GetServerPoolSize возвращает список сохраненных серверов
func (s *stickySessionServerPool) GetBackends() []Backend {
	return s.backends
}

// AddBackend сохраняет сервер в список серверов
func (s *stickySessionServerPool) AddBackend(b Backend) {
	s.backends = append(s.backends, b)
}

// GetServerPoolSize возвращает количество сохраненных серверов
func (s *stickySessionServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
