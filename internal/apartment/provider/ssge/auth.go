package ssge

import (
	"fmt"
	"log/slog"
	"net/http"
)

var (
	refreshTokenURL = "https://home.ss.ge/api/refresh_access_token" //nolint:gosec

	refreshHeaders = map[string]string{
		"accept":             "*/*",
		"accept-language":    "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7",
		"authority":          "home.ss.ge",
		"content-length":     "0",
		"origin":             "https://home.ss.ge",
		"referer":            "https://home.ss.ge/ru/%D0%BD%D0%B5%D0%B4%D0%B2%D0%B8%D0%B6%D0%B8%D0%BC%D0%BE%D1%81%D1%82%D1%8C",
		"sec-ch-ua":          "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Linux\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"user-agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	cookieHeaderTemplate = "ss-session-token=%s;"
	cookieTokenName      = "ss-session-token"
)

func (s *ssge) refreshToken() error {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, refreshTokenURL, nil)
	if err != nil {
		return err
	}

	s.addRefreshHeaders(req)

	s.requestMutex.Lock()
	res, err := s.client.Do(req)
	if err != nil {
		s.requestMutex.Unlock()
		return err
	}
	s.requestMutex.Unlock()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", res.StatusCode)
	}

	s.requestMutex.Lock()
	defer s.requestMutex.Unlock()

	for _, c := range res.Cookies() {
		if c.Name == cookieTokenName {
			s.token = c.Value

			slog.Info("refresh token", "token", s.token)
			return nil
		}
	}

	return fmt.Errorf("not found cookie")
}

func (s *ssge) addRefreshHeaders(req *http.Request) {
	for k, v := range refreshHeaders {
		req.Header.Set(k, v)
	}

	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()

	if len(s.token) > 0 {
		req.Header.Set("cookie", fmt.Sprintf(cookieHeaderTemplate, s.token))
	}
}
