package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Pizhlo/balancer/internal/target/logger"
)

type Targeter interface {
	GetAddress(ctx context.Context) (string, error)
	UpdateStatus(ctx context.Context, status bool, address string) error
}

type Service struct {
	Targeter       Targeter
	counter        int
	mutex          sync.RWMutex
	sleepDuration  time.Duration
	tickerDuration time.Duration
	logger         *log.Logger
}

func New(t Targeter, tickerDuration time.Duration, sleepDuration time.Duration) *Service {
	s := &Service{Targeter: t, tickerDuration: tickerDuration, sleepDuration: sleepDuration}

	return s
}

// CreateLogger создает логгер, который ведет запись в файл, а также запускает тикер
func (s *Service) CreateLogger(address string, strategy string) {
	l, err := logger.New(address, strategy)
	if err != nil {
		log.Fatal("unable to create logger: ", err)
	}

	s.logger = l

	done := make(chan bool)
	go s.startTicker(s.tickerDuration, done)
}

// Handle увеличивает количество текущих запросов, а затем уменьшает
func (s *Service) Handle() {
	s.increment()
	time.Sleep(s.sleepDuration)
	s.decrement()
}

// GetCount возвращает количество текущих запросов
func (s *Service) GetCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.counter
}

func (s *Service) increment() {
	s.mutex.Lock()
	s.counter++
	s.mutex.Unlock()
}

func (s *Service) decrement() {
	s.mutex.Lock()
	s.counter--
	s.mutex.Unlock()
}

func (s *Service) log() {
	log.Println("current number of requests:", s.counter)
	s.logger.Println("current number of requests:", s.counter)
}

func (s *Service) startTicker(d time.Duration, done chan bool) *time.Ticker {
	tick := time.NewTicker(d)

	for {
		select {
		case <-tick.C:
			s.log()
		case <-done:
			break
		default:
		}
	}
}
