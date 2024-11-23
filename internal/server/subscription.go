package server

import (
	"context"
	"log/slog"

	"github.com/irbgeo/apartment-bot/internal/utils"
)

func (s *service) Subscribe(ctx context.Context) <-chan Apartment {
	var id int64
	utils.UnpackVar(ctx, utils.IDKey, &id) // nolint: errcheck

	subCh := make(chan Apartment, 1)

	s.subscribers.Store(id, subCh)

	slog.Info("new subscriber", "id", id)

	return subCh
}

func (s *service) Unsubscribe(ctx context.Context) {
	var id int64
	utils.UnpackVar(ctx, utils.IDKey, &id) // nolint: errcheck

	subCh, isExist := s.subscribers.Load(id)
	if isExist {
		s.subscribers.Delete(id)
		close(subCh.(chan Apartment)) // nolint: errcheck
	}

	slog.Info("unsubscribed", "id", id)
}

func (s *service) sendToSubscribers(a Apartment) {
	s.subscribers.Range(
		func(_, value any) bool {
			subCh := value.(chan Apartment) // nolint: errcheck

			select {
			case subCh <- a:
			default:
			}

			return true
		},
	)
}
