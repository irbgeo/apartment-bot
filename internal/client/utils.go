package client

import (
	"context"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var activeFilter func(ctx context.Context, u *server.User) (*server.Filter, error)

type req interface {
	GetUserID() int64
	SetActiveFilter(i *server.Filter)
}

// WithActiveFilter sets the active filter for the request and then executes the handler.
func WithActiveFilter[REQ req](ctx context.Context, r REQ, handler func(context.Context, REQ) (*server.Filter, error)) (*server.Filter, error) {
	f, err := activeFilter(ctx, &server.User{ID: r.GetUserID()})
	if err != nil {
		return nil, ErrActiveFilterNotFound
	}
	r.SetActiveFilter(f)

	return handler(ctx, r)
}

// StopReceiveHistoryFilter stops receiving history for a specific filter.
func (s *service) StopReceiveHistoryFilter(filterID string) {
	cancel, ok := s.storage.historyReceiving.Load(filterID)
	if !ok {
		return
	}

	cancel.(context.CancelFunc)() // nolint: errcheck
	s.storage.historyReceiving.Delete(filterID)
}

// checkFilter checks if the provided filter is valid.
func checkFilter(f *server.Filter) error {
	if f.Name == nil {
		return ErrUnknownFilterName
	}
	if !f.IsUpdate {
		return errFilterNotChanged
	}

	if f.MinPrice != nil && f.MaxPrice != nil && (*f.MinPrice > *f.MaxPrice) {
		return errMinPriceMoreThanMaxPrice
	}

	if f.MinRooms != nil && f.MaxRooms != nil && (*f.MinRooms > *f.MaxRooms) {
		return errMinRoomsMoreThanMaxRooms
	}

	if f.MinArea != nil && f.MaxArea != nil && (*f.MinArea > *f.MaxArea) {
		return errMinAreaMoreThanMaxArea
	}

	return nil
}

// inSlice checks if a string exists in a slice of strings.
func inSlice(s string, strs []string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}
