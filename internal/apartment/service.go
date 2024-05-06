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

	maxFetchPages int64
	apartmentTTL  time.Duration

	provider provider
}

//go:generate mockery --name provider --structname Provider
type provider interface {
	Apartments(ctx context.Context, page int64) ([]server.Apartment, error)
	IsAvailable(ctx context.Context, a server.Apartment) (bool, error)
	SetInCache(a server.Apartment)
	DeleteFromCache(a server.Apartment)
}

func NewService(
	maxFetchPages int64,
	apartmentTTL time.Duration,
	provider provider,
) *service {
	svc := &service{
		maxFetchPages: maxFetchPages,
		apartmentTTL:  apartmentTTL,

		apartmentCh: make(chan server.Apartment),
		provider:    provider,
	}

	svc.ctx, svc.cancel = context.WithCancel(context.Background())
	return svc
}

func (s *service) Start(opts StartOpts) error {
	go func() {
		updatedTicker := time.NewTicker(opts.UpdateInterval)
		defer updatedTicker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-updatedTicker.C:
				s.update()
			}
		}
	}()

	return nil
}

func (s *service) Stop() {
	s.cancel()
}

func (s *service) Watcher() <-chan server.Apartment {
	return s.apartmentCh
}

func (s *service) IsAvailable(ctx context.Context, a server.Apartment) (bool, error) {
	return s.provider.IsAvailable(ctx, a)
}

func (s *service) SetInCache(a server.Apartment) {
	s.provider.SetInCache(a)
}

func (s *service) DeleteFromCache(a server.Apartment) {
	s.provider.DeleteFromCache(a)
}

func (s *service) Apartments() (<-chan server.Apartment, error) {
	resultCh := make(chan server.Apartment)

	wg := &sync.WaitGroup{}
	go func() {
		var page int64 = 1
		for ; page <= s.maxFetchPages; page++ {
			apartments, err := s.provider.Apartments(s.ctx, page)
			if err != nil {
				slog.Error("fetch apartments", "err", err)
				continue
			}

			for _, a := range apartments {
				wg.Add(1)
				go func(a server.Apartment) {
					resultCh <- a
					wg.Done()
				}(a)
			}

			if len(apartments) == 0 {
				break
			}
		}

		wg.Wait()
		close(resultCh)
	}()

	return resultCh, nil
}

func (s *service) update() {
	var page int64 = 1
	for ; page <= s.maxFetchPages; page++ {
		apartments, err := s.provider.Apartments(s.ctx, page)
		if err != nil {
			slog.Error("fetch apartments", "err", err)
			continue
		}

		for _, a := range apartments {
			go func(a server.Apartment) {
				s.apartmentCh <- a
			}(a)
		}

		if len(apartments) == 0 {
			return
		}

		slog.Info("fetch apartments", "page", page, "count", len(apartments))
	}
}
