package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/irbgeo/apartment-bot/internal/utils"
)

var (
	checkSavedApartmentInterval = 24 * time.Hour
)

type service struct {
	ctx    context.Context
	cancel context.CancelFunc

	apartment apartment
	storage   storage
	filter    filter

	historySending sync.Map
	subscribers    sync.Map

	cityMutex sync.RWMutex
	cities    map[string][]string
}

//go:generate mockery --name apartment --structname Apartment
type apartment interface {
	Watcher() <-chan Apartment
	Apartments() (<-chan Apartment, error)
	IsAvailable(ctx context.Context, apartment Apartment) (bool, error)
	DeleteFromCache(a Apartment)
}

//go:generate mockery --name storage --structname Storage
type storage interface {
	SaveApartment(ctx context.Context, a Apartment) error
	UpdateApartment(ctx context.Context, a Apartment) error
	Apartment(ctx context.Context, f Filter) (<-chan Apartment, error)
	ApartmentCount(ctx context.Context, f Filter) (int64, error)
	DeleteApartment(ctx context.Context, a Apartment) error
	DeleteApartments(ctx context.Context) error

	InsertUser(ctx context.Context, u User) error
	User(ctx context.Context, f Filter) (User, error)
	DeleteUser(ctx context.Context, u User) error

	SaveCity(ctx context.Context, c City) error
	Cities(ctx context.Context) ([]City, error)
}

//go:generate mockery --name filter --structname Filter
type filter interface {
	Add(ctx context.Context, f Filter) (*Filter, error)
	Check(ctx context.Context, a *Apartment)
	Get(ctx context.Context, f Filter) (*Filter, error)
	GetForUser(ctx context.Context, u int64) ([]Filter, error)
	Delete(ctx context.Context, f Filter) error
}

func NewService(
	a apartment,
	s storage,
	f filter,
) *service {
	svc := &service{
		apartment: a,
		storage:   s,
		filter:    f,
	}

	svc.ctx, svc.cancel = context.WithCancel(context.Background())

	return svc
}

func (s *service) Start() error {
	err := s.updatedCities()
	if err != nil {
		return err
	}

	go func() {
		err := s.checkSavedApartment(s.ctx)
		if err != nil {
			slog.Error("check saved apartment", "err", err)
		}
	}()

	go func() {
		checkTicker := time.NewTicker(checkSavedApartmentInterval)
		defer checkTicker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return

			case a, ok := <-s.apartment.Watcher():
				if !ok {
					return
				}

				a, updated := s.saveApartment(a)
				if updated {
					continue
				}

				s.filter.Check(s.ctx, &a)
				if len(a.Filter) == 0 {
					continue
				}

				go func(a Apartment) {
					select {
					case <-s.ctx.Done():
					default:
						s.sendToSubscribers(a)
					}
				}(a)
			case <-checkTicker.C:
				err := s.checkSavedApartment(s.ctx)
				if err != nil {
					slog.Error("check saved apartment", "err", err)
				}

				err = s.updatedCities()
				if err != nil {
					slog.Error("update cities", "err", err)
				}
			}
		}
	}()

	return nil
}

func (s *service) RefreshApartments() error {
	err := s.storage.DeleteApartments(s.ctx)
	if err != nil {
		return err
	}

	newApartmentCh, err := s.apartment.Apartments()
	if err != nil {
		return err
	}

	for a := range newApartmentCh {
		err := s.storage.SaveApartment(s.ctx, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Stop() {
	s.cancel()
	s.subscribers.Range(func(_, value any) bool {
		close(value.(chan Apartment))
		return true
	})
}

func (s *service) SaveFilter(ctx context.Context, f Filter) (int64, error) {
	s.ConnectUser(ctx, User{ID: f.User.ID}) // nolint: errcheck

	filter, err := s.checkFilter(ctx, &f)
	if err != nil {
		return 0, err
	}

	s.stopSendHistoryData(*filter)

	if filter.PauseTimestamp != nil {
		return 0, nil
	}

	count, err := s.storage.ApartmentCount(ctx, *filter)
	if err != nil {
		return 0, err
	}

	_, err = s.filter.Add(ctx, *filter)
	if err != nil {
		return 0, err
	}

	return count, err
}

func (s *service) checkFilter(ctx context.Context, f *Filter) (*Filter, error) {
	existingFilters, err := s.filter.GetForUser(ctx, f.User.ID)
	if err != nil {
		return nil, err
	}

	if len(existingFilters) > 0 {
		if existingFilters[0].ID == f.ID {
			return f, nil
		}

		user, err := s.storage.User(ctx, *f)
		if err != nil {
			return nil, err
		}

		if !user.IsSuperuser {
			return nil, errLimitExceeded
		}
	}

	filter, err := s.filter.Add(ctx, *f)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func (s *service) Filter(ctx context.Context, f Filter) (*Filter, error) {
	filter, err := s.filter.Get(ctx, f)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func (s *service) Filters(ctx context.Context, u User) ([]Filter, error) {
	return s.filter.GetForUser(ctx, u.ID)
}

func (s *service) DeleteFilter(ctx context.Context, f Filter) error {
	s.stopSendHistoryData(f)

	return s.filter.Delete(ctx, f)
}

func (s *service) ConnectUser(ctx context.Context, u User) error {
	f := Filter{User: &User{ID: u.ID}}
	user, _ := s.storage.User(ctx, f)
	user.ID = u.ID

	utils.UnpackVar(ctx, utils.IDKey, &user.ClientID) // nolint: errcheck
	return s.storage.InsertUser(ctx, user)
}

func (s *service) DisconnectUser(ctx context.Context, u User) error {
	f := Filter{User: &User{ID: u.ID}}

	s.stopSendHistoryData(f)

	err := s.filter.Delete(ctx, f)
	if err != nil {
		return err
	}

	return s.storage.DeleteUser(ctx, u)
}

func (s *service) Cities(ctx context.Context) ([]City, error) {
	return s.storage.Cities(ctx)
}

func (s *service) Apartments(ctx context.Context, f Filter) (<-chan Apartment, error) {
	s.stopSendHistoryData(f)

	filter, err := s.filter.Get(ctx, f)
	if err != nil {
		return nil, err
	}

	apartmentCh, err := s.storage.Apartment(ctx, *filter) // nolint: errcheck
	if err != nil {
		return nil, err
	}

	resultCh := s.startSendHistoryData(ctx, *filter, apartmentCh)
	return resultCh, nil
}

func (s *service) checkSavedApartment(ctx context.Context) error {
	apartmentCh, err := s.storage.Apartment(ctx, Filter{})
	if err != nil {
		return fmt.Errorf("get apartments: %w", err)
	}

	for apartment := range apartmentCh {
		s.checkApartment(s.ctx, apartment)
	}
	return nil
}

// saveApartment returns true if apartment was updated
func (s *service) saveApartment(a Apartment) (Apartment, bool) {
	var ok bool
	a.District, ok = s.district(a)

	if !ok {
		city := City{
			Name: a.City,
		}
		if a.District != "" {
			city.District = map[string]struct{}{a.District: {}}
		}

		err := s.storage.SaveCity(s.ctx, city)
		if err != nil {
			slog.Error("save city", "err", err)
		}
	}

	err := s.storage.SaveApartment(s.ctx, a)
	if err != nil {
		_ = s.storage.UpdateApartment(s.ctx, a) // nolint: errcheck
		return a, true
	}

	return a, false
}

func (s *service) district(a Apartment) (string, bool) {
	districts, ok := s.cities[a.City]
	if ok {
		for _, district := range districts {
			if strings.Contains(a.District, district) {
				return district, true
			}
		}
	}
	return a.District, false
}

func (s *service) checkApartment(ctx context.Context, a Apartment) bool {
	isActive, err := s.apartment.IsAvailable(ctx, a)
	if err != nil && err != context.Canceled {
		slog.Error("check_apartment", "err", err)
		return true
	}

	if isActive {
		return true
	}

	s.apartment.DeleteFromCache(a)
	err = s.storage.DeleteApartment(ctx, a)
	if err != nil {
		slog.Error("delete_apartment", "err", err)
	}
	slog.Info("delete_apartment", "url", a.URL, "data", a.OrderDate)
	return false
}

func (s *service) updatedCities() error {
	cities, err := s.storage.Cities(s.ctx)
	if err != nil {
		return err
	}

	s.cityMutex.Lock()
	defer s.cityMutex.Unlock()

	s.cities = make(map[string][]string, len(cities))
	for _, city := range cities {
		s.cities[city.Name] = make([]string, 0, len(city.District))
		for district := range city.District {
			s.cities[city.Name] = append(s.cities[city.Name], district)
		}
	}
	return nil
}
