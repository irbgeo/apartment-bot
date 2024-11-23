package client

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	updateCityInterval           = time.Hour
	checkTurnedOffFilterInterval = time.Hour
	turnedOffFilterTime          = 30 * time.Minute
)

var (
	onceClientFlag int64
)

type service struct {
	ctx      context.Context
	cancel   context.CancelFunc
	srv      srv
	channels channels
	storage  storage
}

type srv interface {
	SaveFilter(context.Context, server.Filter) (int64, error)
	Filter(context.Context, server.Filter) (*server.Filter, error)
	Filters(context.Context, server.User) ([]server.Filter, error)
	DeleteFilter(context.Context, server.Filter) error
	ConnectUser(context.Context, server.User) error
	DisconnectUser(context.Context, server.User) error
	Cities(ctx context.Context) (map[string][]string, error)
	Apartments(ctx context.Context, f server.Filter) (<-chan server.Apartment, <-chan error, error)
	Connect(context.Context) (<-chan server.Apartment, <-chan error, error)
}

func NewService(srv srv, firstCities []string) (*service, error) {
	if !atomic.CompareAndSwapInt64(&onceClientFlag, 0, 1) {
		return nil, errClientAlreadyExist
	}

	ctx, cancel := context.WithCancel(context.Background())
	svc := &service{
		ctx:    ctx,
		cancel: cancel,
		srv:    srv,
		channels: channels{
			apartment: make(chan server.Apartment),
		},
		storage: storage{
			turnedOff: turnedOffStorage{
				filters: make(map[int64]map[string]time.Time),
			},
			cities: citiesStorage{
				firstCities: firstCities,
			},
		},
	}

	activeFilter = svc.ActiveFilter
	return svc, nil
}

func (s *service) Start() error {
	apartmentCh, errCh, err := s.srv.Connect(s.ctx)
	if err != nil {
		return err
	}

	go s.handleApartments(apartmentCh, errCh) // nolint: errcheck

	if err := s.startUpdatingCity(); err != nil {
		return err
	}

	if err := s.startCheckTurnedOffFilters(); err != nil {
		return err
	}

	return nil
}

func (s *service) Stop() {
	s.cancel()
	close(s.channels.apartment)
	atomic.StoreInt64(&onceClientFlag, 0)
}

func (s *service) Watcher() <-chan server.Apartment {
	return s.channels.apartment
}

func (s *service) StartChat(ctx context.Context, u *server.User) error {
	s.storage.disconnectedUsers.Delete(u.ID)
	return s.srv.ConnectUser(ctx, *u)
}

func (s *service) Filters(ctx context.Context, u *server.User) ([]server.Filter, error) {
	return s.srv.Filters(ctx, *u)
}

func (s *service) ActiveFilter(ctx context.Context, u *server.User) (*server.Filter, error) {
	f, exists := s.storage.active.Load(u.ID)
	if !exists {
		return nil, ErrFilterNotFound
	}
	return f.(*server.Filter), nil // nolint: errcheck
}

func (s *service) Filter(ctx context.Context, f *server.Filter) (*server.Filter, error) {
	filter, err := s.srv.Filter(ctx, *f)
	if err != nil {
		return nil, err
	}

	s.storage.active.Store(filter.User.ID, filter)
	return filter, nil
}

func (s *service) StartCreatingFilter(ctx context.Context, u *server.User) *server.Filter {
	defaultAdType := server.RentAdType
	filter := &server.Filter{
		User: &server.User{
			ID: u.ID,
		},
		District: make(map[string]struct{}),
		AdType:   &defaultAdType,
	}

	s.storage.active.Store(u.ID, filter)
	return filter
}

func (s *service) CancelCreatingFilter(_ context.Context, u *server.User) {
	s.storage.active.Delete(u.ID)
}

func (s *service) SaveFilter(ctx context.Context, info *SaveFilterInfo) (*server.Filter, int64, error) {
	activeFilter, err := activeFilter(ctx, info.User)
	if err != nil {
		return nil, 0, err
	}

	if err := checkFilter(activeFilter); err != nil {
		return nil, 0, err
	}

	s.StopReceiveHistoryFilter(activeFilter.ID)

	count, err := s.srv.SaveFilter(ctx, *activeFilter)
	if err != nil {
		return nil, 0, err
	}

	s.handleFilterStatus(activeFilter)
	s.storage.active.Delete(info.User.ID)

	return activeFilter, count, nil
}

func (s *service) handleFilterStatus(f *server.Filter) {
	if f.PauseTimestamp != nil {
		s.turnOffFilter(f)
	} else {
		s.turnOnFilter(f)
	}
}

func (s *service) Apartments(ctx context.Context, f *server.Filter) error {
	s.StopReceiveHistoryFilter(f.ID)

	aCtx, cancel := context.WithCancel(ctx)
	s.storage.historyReceiving.Store(f.ID, cancel)

	apartmentCh, errCh, err := s.srv.Apartments(aCtx, *f)
	if err != nil {
		s.StopReceiveHistoryFilter(f.ID)
		return err
	}

	return s.handleApartments(apartmentCh, errCh)
}

func (s *service) DeleteFilter(ctx context.Context, f *server.Filter) error {
	if len(f.ID) == 0 {
		s.storage.active.Delete(f.User.ID)
		return nil
	}

	s.StopReceiveHistoryFilter(f.ID)

	if err := s.srv.DeleteFilter(ctx, *f); err != nil {
		return err
	}

	s.turnOffFilter(f)
	s.storage.active.Delete(f.User.ID)
	return nil
}

func (s *service) IsAllow(userID int64) bool {
	_, isDisconnected := s.storage.disconnectedUsers.Load(userID)
	return !isDisconnected
}

func (s *service) WorkingFilters(userID int64, filters []string) []string {
	result := make([]string, 0, len(filters))
	for _, name := range filters {
		filter := &server.Filter{
			User: &server.User{ID: userID},
			Name: &name,
		}
		if !s.isTurnedOff(filter) {
			result = append(result, name)
		}
	}
	return result
}

func (s *service) AvailableCities() []string {
	cities := make([]string, 0)
	tempCities := make(map[string]struct{})

	s.storage.cities.Range(func(key, _ interface{}) bool {
		tempCities[key.(string)] = struct{}{} // nolint: errcheck
		return true
	})

	// Add first cities in order
	for _, city := range s.storage.cities.firstCities {
		cities = append(cities, city)
		delete(tempCities, city)
	}

	// Add remaining cities
	for city := range tempCities {
		cities = append(cities, city)
	}

	return cities
}

func (s *service) AvailableDistrictsForCity(city string) []string {
	districts, ok := s.storage.cities.Load(city)
	if ok {
		return nil
	}
	return districts.([]string) // nolint: errcheck
}

func (s *service) startUpdatingCity() error {
	if err := s.updateCity(); err != nil {
		return err
	}

	go s.cityUpdateLoop()
	return nil
}

func (s *service) cityUpdateLoop() {
	ticker := time.NewTicker(updateCityInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.updateCity(); err != nil {
				slog.Error("update city failed", "error", err)
			}
		}
	}
}

func (s *service) updateCity() error {
	cities, err := s.srv.Cities(s.ctx)
	if err != nil {
		return err
	}

	for name, districts := range cities {
		s.storage.cities.Store(name, districts)
	}
	return nil
}

func (s *service) startCheckTurnedOffFilters() error {
	go s.checkTurnedOffFiltersLoop()
	return nil
}

func (s *service) checkTurnedOffFiltersLoop() {
	ticker := time.NewTicker(checkTurnedOffFilterInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkTurnedOffFilters()
		}
	}
}

func (s *service) checkTurnedOffFilters() {
	s.storage.turnedOff.RLock()
	defer s.storage.turnedOff.RUnlock()

	now := time.Now()
	for userID, filters := range s.storage.turnedOff.filters {
		for filterID, t := range filters {
			if now.Sub(t) > turnedOffFilterTime {
				delete(s.storage.turnedOff.filters[userID], filterID)
			}
		}
		if len(s.storage.turnedOff.filters[userID]) == 0 {
			delete(s.storage.turnedOff.filters, userID)
		}
	}
}

func (s *service) turnOffFilter(f *server.Filter) {
	s.storage.turnedOff.Lock()
	defer s.storage.turnedOff.Unlock()

	if _, exists := s.storage.turnedOff.filters[f.User.ID]; !exists {
		s.storage.turnedOff.filters[f.User.ID] = make(map[string]time.Time)
	}
	s.storage.turnedOff.filters[f.User.ID][f.ID] = time.Now()
}

func (s *service) turnOnFilter(f *server.Filter) {
	s.storage.turnedOff.Lock()
	defer s.storage.turnedOff.Unlock()

	if filters, exists := s.storage.turnedOff.filters[f.User.ID]; exists {
		delete(filters, f.ID)
		if len(filters) == 0 {
			delete(s.storage.turnedOff.filters, f.User.ID)
		}
	}
}

func (s *service) isTurnedOff(f *server.Filter) bool {
	s.storage.turnedOff.RLock()
	defer s.storage.turnedOff.RUnlock()

	filters, exists := s.storage.turnedOff.filters[f.User.ID]
	if !exists {
		return false
	}

	if _, isTurnedOff := filters[f.ID]; isTurnedOff {
		filters[f.ID] = time.Now()
		return true
	}
	return false
}

func (s *service) handleApartments(apartmentCh <-chan server.Apartment, errCh <-chan error) error {
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case err, ok := <-errCh:
			if !ok {
				return nil
			}
			return err
		case apt, ok := <-apartmentCh:
			if !ok {
				return nil
			}
			slog.Info("received apartment",
				"id", apt.ID,
				"filter", apt.Filter,
			)
			s.channels.apartment <- apt
		}
	}
}
