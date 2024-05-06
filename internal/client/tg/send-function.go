package tg

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

func (s *service) sendApartment(a server.Apartment) {
	for userID, filters := range a.Filter {
		filters = s.service.WorkingFilters(userID, filters)
		if len(filters) == 0 {
			continue
		}

		messageCount, apartmentAlbum := s.apartmentMessage(a, filters)
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		var err error

		switch messageCount {
		case 0:
			err = s.sendSingleApartmentMessage(userID, a, filters)
		default:
			err = s.sendApartmentAlbum(userID, apartmentAlbum)
		}

		if err != nil {
			s.handleError(userID, err)
		}
	}
}

func (s *service) sendSingleApartmentMessage(userID int64, a server.Apartment, filters []string) error {
	var err error
	if !s.service.IsAllow(userID) {
		return nil
	}
	defer func() {
		interval := s.messageSendInterval
		if err != nil {
			interval = s.extractRetryTime(err.Error())
			slog.Info("sendApartmentAlbum", "interval", interval)
		}
		<-time.After(interval)
	}()

	_, err = s.b.Send(&tele.User{ID: userID}, apartmentString(a, filters))
	return err
}

func (s *service) sendApartmentAlbum(userID int64, apartmentAlbum tele.Album) error {
	var err error
	if !s.service.IsAllow(userID) {
		return nil
	}
	defer func() {
		interval := s.messageSendInterval
		if err != nil {
			interval = s.extractRetryTime(err.Error())
			slog.Info("sendApartmentAlbum", "interval", interval)
		}
		<-time.After(interval)
	}()

	_, err = s.b.SendAlbum(&tele.User{ID: userID}, apartmentAlbum)
	return err
}

func (s *service) apartmentMessage(a server.Apartment, filters []string) (int, tele.Album) {
	var (
		resultAlbum  tele.Album
		messageCount int
		overallSize  float64
	)
	for _, photoURL := range a.PhotoURLs {
		size, err := getImageSizeMB(photoURL)
		if err != nil {
			continue
		}

		overallSize += size
		if overallSize >= maxImageSizeMB {
			break
		}

		photoURL = replaceURLSingleSlash(photoURL)

		photo := &tele.Photo{
			File: tele.FromURL(photoURL),
		}

		if messageCount == 0 {
			photo.Caption = apartmentString(a, filters)
		}
		resultAlbum = append(resultAlbum, photo)
		messageCount++

		if messageCount == s.maxPhotoCount {
			break
		}
	}

	return messageCount, resultAlbum
}

func (s *service) sendPinMessage(msg Message) {
	if !s.service.IsAllow(msg.UserID) {
		return
	}

	m, err := s.b.Send(&tele.User{ID: msg.UserID}, msg.Text)
	if err != nil {
		s.handleError(msg.UserID, err)
		return
	}

	err = s.b.Pin(m)
	if err != nil {
		s.handleError(msg.UserID, err)
	}
}

func (s *service) sendSettingFilter(c tele.Context, f *server.Filter) error {
	values := getValue(c)

	settingPageIdx := 0
	if len(values) != 0 && values[0] == nextPage {
		settingPageIdx, _ = strconv.Atoi(values[1])
	}

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        s.filterString(f) + "\n\n📝 You can change:",
		ReplyMarkup: s.filterSettingMarkup(f, settingPageIdx),
	}

	return s.sendMessage(msg, settingFilterMessage)
}

func (s *service) sendSavedFilter(c tele.Context, count int64, f *server.Filter) error {
	userID := c.Sender().ID

	err := s.messages.Clean(userID)
	if err != nil {
		return err
	}

	msg := s.filterString(f)
	var markup *tele.ReplyMarkup
	if f.PauseTimestamp == nil {
		markup = filterSavedMarkup(f.ID, count)
	}

	m, err := s.b.Send(c.Sender(), msg, markup)
	if err != nil {
		return err
	}

	s.messages.Store(userID, m, botMessage)

	return nil
}

func (s *service) sendMessage(msg *tele.Message, messageType MessageType) error {
	m, isExist, err := s.messages.GetOrCleanTill(msg.Sender.ID, messageType, settingFilterMessage)
	if err != nil {
		return err
	}

	if !isExist {
		m, err := s.b.Send(msg.Sender, msg.Text, msg.ReplyMarkup)
		if err != nil {
			return err
		}

		s.messages.Store(msg.Sender.ID, m, messageType)
		return nil
	}

	m, err = s.b.Edit(m, msg.Text, msg.ReplyMarkup)
	if err != nil {
		return err
	}

	s.messages.Store(msg.Sender.ID, m, messageType)
	return nil
}

var paramsOrder = []string{
	changeName,
	changeAdType,
	changeBuildingStatus,
	changeCity,
	changeDistrict,
	changeMinPrice,
	changeMaxPrice,
	changeMinRooms,
	changeMaxRooms,
	changeMinArea,
	changeMaxArea,
	changeOwnerType,
	changeLocation,
	changeMaxDistance,
}

func (s *service) filterString(f *server.Filter) string {
	var parts []string
	for _, name := range paramsOrder {
		p := s.params[name]
		if p.toString != nil {
			parts = append(parts, p.toString(f))
		}
	}

	return strings.Join(parts, "\n")
}

func rangeStr(min, max *float64) string {
	switch {
	case min != nil && max != nil:
		return fmt.Sprintf("%0.0f - %0.0f", *min, *max)
	case min != nil:
		return fmt.Sprintf("%0.0f - ∞", *min)
	case max != nil:
		return fmt.Sprintf("0 - %0.0f", *max)
	}

	return anyValue
}
