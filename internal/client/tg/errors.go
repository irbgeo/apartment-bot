package tg

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	errNotFoundHandler  = errors.New("handler not found")
	errNotFoundLocation = errors.New("location not found\nSend location from Telegram")
)

func (s *service) errorMiddleware(h tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if err := h(c); err != nil {
			if err := s.handleTelegramError(c, err); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *service) handleTelegramError(c tele.Context, err error) error {
	s.service.ErrorHandler(s.ctx, userFromContext(c), err)
	if err == client.ErrActiveFilterNotFound {
		s.userAction.Delete(c.Sender().ID)
	}
	if err := s.sendErrorMessage(c, err); err != nil {
		return err
	}

	return nil
}

func (s *service) sendErrorMessage(c tele.Context, err error) error {
	userID := c.Sender().ID

	msg := "ERROR: " + err.Error() + "\n/help"
	if err.Error() == client.ErrActiveFilterNotFound.Error() {
		msg = fmt.Sprintf(notActiveFilterMessageLayout, filterCommand)
	}

	m, err := s.b.Send(c.Sender(), msg)
	if err != nil {
		return err
	}
	s.messages.Store(userID, m, errMessage)

	if actionType, isExist := s.userAction.Load(userID); isExist {
		if err := s.params[actionType.(string)].init(c); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) handleError(userID int64, err error) {
	user := &server.User{ID: userID}

	switch extractCode(err.Error()) {
	case 403:
		s.service.BlockErrorHandler(s.ctx, user, err)
		return
	}

	if strings.Contains(err.Error(), "USER_IS_BLOCKED") {
		s.service.BlockErrorHandler(s.ctx, user, err)
		return
	}

	s.service.ErrorHandler(s.ctx, &server.User{ID: userID}, err)
}

var reCode = regexp.MustCompile(`\(([^)]+)\)`)

func extractCode(input string) int {
	matches := reCode.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0
	}

	code, _ := strconv.Atoi(matches[1])
	return code
}

var reTimeout = regexp.MustCompile(`retry after (\d+)`)

func (s *service) extractRetryTime(input string) time.Duration {
	match := reTimeout.FindStringSubmatch(input)
	if len(match) > 1 {
		timeout, _ := strconv.Atoi(match[1])
		return time.Duration(timeout) * time.Second
	}
	return s.messageSendInterval
}
