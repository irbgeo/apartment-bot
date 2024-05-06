package filter

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/irbgeo/apartment-bot/internal/server"
)

type filter struct {
	storage storage

	filter sync.Map
}

type storage interface {
	SaveFilter(ctx context.Context, f server.Filter) error
	FilterList(ctx context.Context, f server.Filter) ([]server.Filter, error)
	DeleteFilter(ctx context.Context, f server.Filter) error
}

func New(filterStorage storage) (*filter, error) {
	f := &filter{
		storage: filterStorage,
	}

	filterList, err := f.storage.FilterList(context.Background(), server.Filter{})
	if err != nil {
		return nil, err
	}

	for _, filter := range filterList {
		f.filter.Store(filter.ID, filter)
	}

	return f, nil
}

func (s *filter) Add(ctx context.Context, f server.Filter) (*server.Filter, error) {
	if len(f.ID) == 0 {
		f.ID = uuid.New().String()
	}

	prevFilter, isExist := s.filter.Load(f.ID)
	if isExist {
		f.FromTimestamp = prevFilter.(server.Filter).PauseTimestamp // nolint: errcheck
	}

	if err := s.storage.SaveFilter(ctx, f); err != nil {
		return nil, err
	}

	s.filter.Store(f.ID, f)
	return &f, nil
}

func (s *filter) Check(ctx context.Context, a *server.Apartment) {
	a.Filter = make(map[int64][]string)

	s.filter.Range(
		func(_, value any) bool {
			f := value.(server.Filter) // nolint: errcheck

			if f.IsFit(a) {
				a.Filter[f.User.ID] = append(a.Filter[f.User.ID], *f.Name)
			}
			return true
		},
	)
}

func (s *filter) Get(ctx context.Context, f server.Filter) (*server.Filter, error) {
	filterList, err := s.storage.FilterList(context.Background(), f)
	if err != nil {
		return nil, err
	}

	for _, filter := range filterList {
		s.filter.Store(filter.ID, filter)
	}
	if len(filterList) == 0 {
		return nil, errFilterNotFound
	}

	return &filterList[0], nil
}

func (s *filter) GetForUser(ctx context.Context, id int64) ([]server.Filter, error) {
	filterList, err := s.storage.FilterList(context.Background(), server.Filter{User: &server.User{ID: id}})
	if err != nil {
		return nil, err
	}

	for _, filter := range filterList {
		s.filter.Store(filter.ID, filter)
	}

	return filterList, nil
}

func (s *filter) Delete(ctx context.Context, f server.Filter) error {
	filterList, err := s.storage.FilterList(context.Background(), f)
	if err != nil {
		return err
	}

	for _, f := range filterList {
		if err := s.storage.DeleteFilter(ctx, f); err != nil {
			return err
		}
		s.filter.Delete(f.ID)
	}

	return nil
}
