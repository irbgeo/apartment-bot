package apartment

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

type service struct {
	ctx    context.Context
	cancel context.CancelFunc

	apartmentCh chan server.Apartment
	errCh       chan error

	maxFetchPages int64
	apartmentTTL  time.Duration

	provider provider
	mu       sync.RWMutex
}

// provider represents the data provider interface
type provider interface {
	Apartments(ctx context.Context, page int64) ([]server.Apartment, error)
	IsAvailable(ctx context.Context, a server.Apartment) (bool, error)
	SetInCache(a server.Apartment)
	DeleteFromCache(a server.Apartment)
}

// NewService creates a new apartment service instance
func NewService(
	cfg Config,
	p provider,
) *service {
	ctx, cancel := context.WithCancel(context.Background())

	return &service{
		ctx:           ctx,
		cancel:        cancel,
		maxFetchPages: cfg.MaxFetchPages,
		apartmentTTL:  cfg.ApartmentTTL,
		apartmentCh:   make(chan server.Apartment, 10),
		errCh:         make(chan error, 10),
		provider:      p,
	}
}

func (s *service) Start(updateInterval time.Duration) error {
	go s.startUpdateLoop(updateInterval)
	return nil
}

func (s *service) Stop() {
	s.cancel()
}

func (s *service) Watcher() <-chan server.Apartment {
	return s.apartmentCh
}

func (s *service) IsAvailable(ctx context.Context, a server.Apartment) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.IsAvailable(ctx, a)
}

func (s *service) SetInCache(a server.Apartment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider.SetInCache(a)
}

func (s *service) DeleteFromCache(a server.Apartment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider.DeleteFromCache(a)
}

func (s *service) Apartments() (<-chan server.Apartment, error) {
	resultCh := make(chan server.Apartment, s.maxFetchPages)

	go s.fetchApartments(resultCh)

	return resultCh, nil
}

func (s *service) startUpdateLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.update()
		}
	}
}

func (s *service) fetchApartments(resultCh chan<- server.Apartment) {
	var wg sync.WaitGroup
	defer close(resultCh)

	for page := int64(1); page <= s.maxFetchPages; page++ {
		apartments, err := s.fetchPage(page)
		if err != nil {
			slog.Error("fetch apartments page", "page", page, "error", err)
			continue
		}

		if len(apartments) == 0 {
			break
		}

		wg.Add(len(apartments))
		for _, apt := range apartments {
			go func(a server.Apartment) {
				defer wg.Done()
				resultCh <- a
			}(apt)
		}
	}

	wg.Wait()
}

func (s *service) update() {
	for page := int64(1); page <= s.maxFetchPages; page++ {
		apartments, err := s.fetchPage(page)
		if err != nil {
			slog.Error("update apartments page", "page", page, "error", err)
			continue
		}

		if len(apartments) == 0 {
			return
		}

		s.broadcastApartments(apartments)
		slog.Info("fetched apartments", "page", page, "count", len(apartments))
	}
}

func (s *service) fetchPage(page int64) ([]server.Apartment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Apartments(s.ctx, page)
}

func (s *service) broadcastApartments(apartments []server.Apartment) {
	for _, apt := range apartments {
		go func(a server.Apartment) {
			select {
			case <-s.ctx.Done():
				return
			case s.apartmentCh <- a:
			}
		}(apt)
	}
}
