package client

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	// TODO: remove this import

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
	ctx                  context.Context
	cancel               context.CancelFunc
	srv                  srv
	activeFilter         sync.Map
	turnedOffFilterMutex sync.RWMutex
	turnedOffFilter      map[int64]map[string]time.Time
	historyReceiving     sync.Map
	disconnectedUser     sync.Map
	apartmentCh          chan server.Apartment
	firstCities          []string
	cities               sync.Map
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
	StartApartmentWatcher(context.Context) (<-chan server.Apartment, <-chan error, error)
}

func NewService(srv srv, firstCities []string) (*service, error) {
	if !atomic.CompareAndSwapInt64(&onceClientFlag, 0, 1) {
		return nil, errClientAlreadyExist
	}

	ctx, cancel := context.WithCancel(context.Background())
	svc := &service{
		ctx:             ctx,
		cancel:          cancel,
		srv:             srv,
		firstCities:     firstCities,
		apartmentCh:     make(chan server.Apartment),
		turnedOffFilter: make(map[int64]map[string]time.Time),
	}

	activeFilter = svc.ActiveFilter

	return svc, nil
}

func (s *service) Start() error {
	apartmentCh, errCh, err := s.srv.StartApartmentWatcher(s.ctx)
	if err != nil {
		return err
	}
	go func() {
		if err := s.receivingFromApartmentChan(s.ctx, apartmentCh, errCh); err != nil {
			slog.Error("finish_apartment_watcher", "err", err)
		}
	}()

	err = s.startUpdatingCity(s.ctx)
	if err != nil {
		return err
	}

	err = s.startCheckTurnedOffFilters(s.ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Stop() {
	s.cancel()
	close(s.apartmentCh)
	atomic.StoreInt64(&onceClientFlag, 0)
}

func (s *service) Watcher() <-chan server.Apartment {
	return s.apartmentCh
}

func (s *service) StartChat(ctx context.Context, u *server.User) error {
	s.disconnectedUser.Delete(u.ID)
	return s.srv.ConnectUser(ctx, *u)
}

func (s *service) Filters(ctx context.Context, u *server.User) ([]server.Filter, error) {
	return s.srv.Filters(ctx, *u)
}

func (s *service) ActiveFilter(ctx context.Context, u *server.User) (*server.Filter, error) {
	f, isExist := s.activeFilter.Load(u.ID)
	if !isExist {
		return nil, ErrFilterNotFound
	}

	return f.(*server.Filter), nil
}

func (s *service) Filter(ctx context.Context, f *server.Filter) (*server.Filter, error) {
	f, err := s.srv.Filter(ctx, *f)
	if err != nil {
		return nil, err
	}

	s.activeFilter.Store(f.User.ID, f)
	return f, nil
}

func (s *service) StartCreatingFilter(ctx context.Context, u *server.User) *server.Filter {
	defaultAdType := server.RentAdType

	newFilter := &server.Filter{
		User: &server.User{
			ID: u.ID,
		},
		District: make(map[string]struct{}),
		AdType:   &defaultAdType,
	}

	s.activeFilter.Store(u.ID, newFilter)
	return newFilter
}

func (s *service) CancelCreatingFilter(ctx context.Context, u *server.User) {
	s.activeFilter.Delete(u.ID)
}

func (s *service) SaveFilter(ctx context.Context, i *SaveFilterInfo) (*server.Filter, int64, error) {
	activeFilter, err := activeFilter(ctx, i.User)
	if err != nil {
		return nil, 0, err
	}

	err = checkFilter(activeFilter)
	if err != nil {
		return nil, 0, err
	}

	s.StopReceiveHistoryFilter(activeFilter.ID)

	count, err := s.srv.SaveFilter(ctx, *activeFilter)
	if err != nil {
		return nil, 0, err
	}

	if activeFilter.PauseTimestamp != nil {
		s.turnOffFilter(activeFilter)
	} else {
		s.turnOnFilter(activeFilter)
	}

	s.activeFilter.Delete(i.User.ID)
	return activeFilter, count, nil
}

func (s *service) Apartments(ctx context.Context, f *server.Filter) error {
	s.StopReceiveHistoryFilter(f.ID)

	aCtx, cancel := context.WithCancel(ctx)
	s.historyReceiving.Store(f.ID, cancel)

	apartmentCh, errCh, err := s.srv.Apartments(aCtx, *f)
	if err != nil {
		s.StopReceiveHistoryFilter(f.ID)
		return err
	}

	err = s.receivingFromApartmentChan(aCtx, apartmentCh, errCh)
	if err != nil {
		s.StopReceiveHistoryFilter(f.ID)
		return err
	}

	return nil
}

func (s *service) DeleteFilter(ctx context.Context, f *server.Filter) error {
	if len(f.ID) == 0 {
		s.activeFilter.Delete(f.User.ID)
		return nil
	}

	s.StopReceiveHistoryFilter(f.ID)

	if err := s.srv.DeleteFilter(ctx, *f); err != nil {
		return err
	}
	s.turnOffFilter(f)
	s.activeFilter.Delete(f.User.ID)

	return nil
}

func (s *service) IsAllow(userID int64) bool {
	_, isDisconnected := s.disconnectedUser.Load(userID)

	return !isDisconnected
}

func (s *service) WorkingFilters(userID int64, f []string) []string {
	workingFilters := make([]string, 0, len(f))
	for _, filterName := range f {
		name := filterName

		f := &server.Filter{
			User: &server.User{
				ID: userID,
			},
			Name: &name,
		}
		if !s.isTurnedOff(f) {
			workingFilters = append(workingFilters, filterName)
		}
	}
	return workingFilters
}

func (s *service) AvailableCities() []string {
	availableCities := make([]string, 0)

	var tmpCities sync.Map
	s.cities.Range(func(key, value any) bool {
		tmpCities.Store(key, value)
		return true
	})

	for _, c := range s.firstCities {
		availableCities = append(availableCities, c)
		tmpCities.Delete(c)
	}

	tmpCities.Range(func(key, _ any) bool {
		availableCities = append(availableCities, key.(string))
		return true
	})

	return availableCities
}

func (s *service) AvailableDistrictsForCity(city string) []string {
	districts, _ := s.cities.Load(city)
	return districts.([]string)
}

func (s *service) startUpdatingCity(ctx context.Context) error {
	err := s.updateCity()
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(updateCityInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := s.updateCity()
				if err != nil {
					slog.Error("update_city", "err", err)
				}
			}
		}
	}()
	return nil
}

func (s *service) updateCity() error {
	cities, err := s.srv.Cities(s.ctx)
	if err != nil {
		return err
	}

	for name, districts := range cities {
		s.cities.Store(name, districts)
	}
	return nil
}

func (s *service) startCheckTurnedOffFilters(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(checkTurnedOffFilterInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkTurnedOffFilters()
			}
		}
	}()
	return nil
}

func (s *service) checkTurnedOffFilters() {
	s.turnedOffFilterMutex.RLock()
	defer s.turnedOffFilterMutex.RUnlock()

	for userID, filters := range s.turnedOffFilter {
		for filterID, t := range filters {
			if time.Since(t) > turnedOffFilterTime {
				delete(s.turnedOffFilter[userID], filterID)
			}

			if len(s.turnedOffFilter[userID]) == 0 {
				delete(s.turnedOffFilter, userID)
			}
		}
	}
}

func (s *service) turnOffFilter(f *server.Filter) {
	s.turnedOffFilterMutex.Lock()
	defer s.turnedOffFilterMutex.Unlock()

	if _, ok := s.turnedOffFilter[f.User.ID]; !ok {
		s.turnedOffFilter[f.User.ID] = make(map[string]time.Time)
	}

	s.turnedOffFilter[f.User.ID][f.ID] = time.Now()
}

func (s *service) turnOnFilter(f *server.Filter) {
	s.turnedOffFilterMutex.Lock()
	defer s.turnedOffFilterMutex.Unlock()

	if _, ok := s.turnedOffFilter[f.User.ID]; !ok {
		return
	}

	delete(s.turnedOffFilter[f.User.ID], f.ID)
}

func (s *service) isTurnedOff(f *server.Filter) bool {
	s.turnedOffFilterMutex.RLock()
	defer s.turnedOffFilterMutex.RUnlock()

	if _, ok := s.turnedOffFilter[f.User.ID]; !ok {
		return false
	}

	_, isTurnedOff := s.turnedOffFilter[f.User.ID][f.ID]
	if isTurnedOff {
		s.turnedOffFilter[f.User.ID][f.ID] = time.Now()
	}

	return isTurnedOff
}

func (s *service) receivingFromApartmentChan(ctx context.Context, apartmentCh <-chan server.Apartment, errCh <-chan error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case err, isOpen := <-errCh:
			if !isOpen {
				return nil
			}
			return err
		case a, isOpen := <-apartmentCh:
			if !isOpen {
				return nil
			}

			slog.Info("receive_apartment", "id", a.ID, "server.User", a.Filter)
			s.apartmentCh <- a
		}
	}
}
