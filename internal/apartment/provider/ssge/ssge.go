package ssge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	rentRealEstateDealType int64 = 1
	saleRealEstateDealType int64 = 4
)

var (
	apartmentURLTemplate  = "https://api-gateway.ss.ge/v1/RealEstate/details?applicationId=%d"
	apartmentListURL      = "https://api-gateway.ss.ge/v1/RealEstate/LegendSearch"
	authTokenTemplate     = "Bearer %s"
	apartmentActiveStatus = "active"

	requestTimeout       = 1 * time.Minute
	pageSize       int64 = 16
	apartmentTTL         = 7 * 24 * time.Hour
)

type ssge struct {
	ctx    context.Context
	cancel context.CancelFunc

	requestMutex sync.Mutex
	client       *http.Client

	tokenMutex sync.RWMutex
	token      string

	cacheID sync.Map
}

func NewSSGEProvider() *ssge {
	p := &ssge{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	return p
}

func (s *ssge) Start(refreshTokenInterval time.Duration) error {
	err := s.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		refreshTokenTicker := time.NewTicker(refreshTokenInterval)
		defer refreshTokenTicker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-refreshTokenTicker.C:
				err := s.refreshToken()
				if err != nil {
					slog.Error("refresh token", "err", err)
					continue
				}
			}
		}
	}()

	return nil
}

func (s *ssge) Stop() {
	s.cancel()
}

func (s *ssge) Apartments(ctx context.Context, page int64) ([]server.Apartment, error) {
	requestBody := requestApartmentListBody{
		AdvancedSearch: advancedSearch{
			WithImageOnly: true,
		},
		RealEstateType: 5,
		CurrencyID:     1,
		Order:          1,
		Page:           page,
		PageSize:       pageSize,
	}

	body, _ := json.Marshal(requestBody)

	apartmentsData, err := s.request(ctx, http.MethodPost, apartmentListURL, body)
	if err != nil {
		return nil, err
	}

	apartments := &apartmentList{}
	err = json.Unmarshal(apartmentsData, apartments)
	if err != nil {
		return nil, err
	}

	result := make([]server.Apartment, 0, pageSize)
	for _, apartment := range apartments.Data {
		_, isExist := s.cacheID.Load(apartment.ApplicationID)
		if isExist {
			continue
		}

		a, err := s.apartment(ctx, apartment.ApplicationID)
		if err != nil {
			slog.Error("get apartment", "err", err)
			continue
		}

		if !check(a) {
			continue
		}

		s.cacheID.Store(a.ApplicationID, struct{}{})

		result = append(result, toServerApartment(*a))
	}

	return result, nil
}

func (s *ssge) IsAvailable(ctx context.Context, a server.Apartment) (bool, error) {
	aData, err := s.apartment(ctx, a.ID)
	if err != nil {
		return true, err
	}

	return check(aData), nil
}

func (s *ssge) SetInCache(a server.Apartment) {
	s.cacheID.Store(a.ID, struct{}{})
}

func (s *ssge) DeleteFromCache(a server.Apartment) {
	s.cacheID.Delete(a.ID)
}

func (s *ssge) apartment(ctx context.Context, id int64) (*apartment, error) {
	apartmentData, err := s.request(ctx, http.MethodPut, fmt.Sprintf(apartmentURLTemplate, id), nil)
	if err != nil {
		return nil, err
	}

	a := &apartment{}
	json.Unmarshal(apartmentData, a) // nolint:errcheck

	return a, nil
}

func (s *ssge) request(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	s.addRequestHeaders(req)

	s.requestMutex.Lock()
	res, err := s.client.Do(req)
	if err != nil {
		s.requestMutex.Unlock()
		return nil, err
	}
	s.requestMutex.Unlock()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *ssge) addRequestHeaders(req *http.Request) {
	headers := map[string]string{
		"Accept-Language": "en",
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf(authTokenTemplate, s.token))
}

func check(a *apartment) bool {
	_, ok := dealTypeMap[a.RealEstateDealTypeID]
	return !a.IsInactiveApplication && time.Since(a.OrderDate) < apartmentTTL && ok
}
