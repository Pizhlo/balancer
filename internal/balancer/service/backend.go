package service

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend interface {
	SetAlive(bool)
	IsAlive() bool
	GetURL() *url.URL
	GetActiveConnections() int
	Serve(http.ResponseWriter, *http.Request)
}

type backend struct {
	url          *url.URL
	alive        bool
	mux          sync.RWMutex
	connections  int
	reverseProxy *httputil.ReverseProxy
}

// GetActiveConnections возвращает активные соединения
func (b *backend) GetActiveConnections() int {
	log.Println("GetActiveConnections ", b.url)
	b.mux.RLock()
	connections := b.connections
	b.mux.RUnlock()
	return connections
}

// SetAlive устанавливает статус соединения: активно или нет
func (b *backend) SetAlive(alive bool) {
	log.Printf("setting alive %s, status %v\n\n", b.url, alive)
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

// IsAlive возвращает статус соединения
func (b *backend) IsAlive() bool {
	log.Printf("checking alive %s\n\n", b.url)
	b.mux.RLock()
	alive := b.alive
	defer b.mux.RUnlock()
	return alive
}

// IsAlive возвращает url соединения
func (b *backend) GetURL() *url.URL {
	log.Printf("getting url %s\n\n", b.url)
	return b.url
}

// IsAlive производит перенаправление запроса
func (b *backend) Serve(rw http.ResponseWriter, req *http.Request) {
	log.Println("backend serving; url =", b.url, "connections = ", b.connections)
	defer func() {
		b.mux.Lock()
		b.connections--
		b.mux.Unlock()
	}()

	b.mux.Lock()
	b.connections++
	b.mux.Unlock()

	log.Println("serve: b.url = ", b.url)
	log.Println("backend serving; url =", b.url, "added connections = ", b.connections)
	b.reverseProxy.ServeHTTP(rw, req)
}

// NewBackend создает новую структуру Backend с переданным url и reverseProxy
func NewBackend(u *url.URL, rp *httputil.ReverseProxy) Backend {
	b := &backend{
		url:          u,
		alive:        true,
		reverseProxy: rp,
	}
	log.Printf("created new backend: %+v\n\n", b)
	return b
}

// NewServerPool создает новую структуру в зависимости от указанного алгоритма балансировки
func NewServerPool(strategy string) (ServerPool, error) {
	switch strategy {
	case "round-robin":
		return &roundRobinServerPool{
			backends: make([]Backend, 0),
			current:  0,
		}, nil
	case "least-connections":
		return &lcServerPool{
			backends: make([]Backend, 0),
		}, nil
	default:
		return nil, fmt.Errorf("Invalid strategy")
	}
}
