package client

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	ErrActiveFilterNotFound = errors.New("active filter not found")
	ErrUnknownFilterName    = errors.New("filter name is not set")

	errClientAlreadyExist       = errors.New("client already exist")
	ErrFilterNotFound           = errors.New("filter not found")
	errFilterNotChanged         = errors.New("filter not changed")
	errMinPriceMoreThanMaxPrice = errors.New("min price more than max price")
	errMinRoomsMoreThanMaxRooms = errors.New("min rooms more than max rooms")
	errMinAreaMoreThanMaxArea   = errors.New("min area more than max area")
)

func (s *service) FloodErrorHandler(ctx context.Context, u *server.User, retryAt time.Duration) {
	slog.Error("flood_err", "retry_at", retryAt)
}

func (s *service) BlockErrorHandler(ctx context.Context, u *server.User, err error) {
	slog.Error("block_err", "user_id", u.ID, "err", err)

	filters, _ := s.srv.Filters(ctx, *u)

	for _, f := range filters {
		s.StopReceiveHistoryFilter(f.ID)
	}

	if err := s.srv.DisconnectUser(ctx, *u); err != nil {
		slog.Error("disconnect_user", "user_id", u.ID, "err", err)
	}
}

func (s *service) ErrorHandler(ctx context.Context, u *server.User, err error) {
	slog.Error("error_handler", "user_id", u.ID, "err", err)
}
