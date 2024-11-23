package tg

import (
	"fmt"
	"strconv"
	"strings"

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

		if !s.service.IsAllow(userID) {
			return
		}

		var (
			err     error
			message any
		)

		switch messageCount {
		case 0:
			message = apartmentString(a, filters)
		default:
			message = apartmentAlbum
		}

		_, err = s.sendMessageToBot(userID, message)
		if err != nil {
			s.handleError(userID, err)
		}
	}
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

	err := s.messages.CleanUserMessages(userID)
	if err != nil {
		return err
	}

	msg := s.filterString(f)
	var markup *tele.ReplyMarkup
	if f.PauseTimestamp == nil {
		markup = filterSavedMarkup(f.ID, count)
	}

	m, err := s.sendMessageToBot(userID, msg, markup)
	if err != nil {
		return err
	}

	s.messages.StoreMessage(userID, m, botMessage)

	return nil
}

func (s *service) sendMessage(msg *tele.Message, messageType MessageType) error {
	m, isExist, err := s.messages.GetOrCleanTill(msg.Sender.ID, messageType, settingFilterMessage)
	if err != nil {
		return err
	}

	if !isExist {
		m, err := s.sendMessageToBot(msg.Sender.ID, msg.Text, msg.ReplyMarkup)
		if err != nil {
			return err
		}

		s.messages.StoreMessage(msg.Sender.ID, m, messageType)
		return nil
	}

	m, err = s.b.Edit(m, msg.Text, msg.ReplyMarkup)
	if err != nil {
		return err
	}

	s.messages.StoreMessage(msg.Sender.ID, m, messageType)
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

func rangeStr(minValue, maxValue *float64) string {
	switch {
	case minValue != nil && maxValue != nil:
		return fmt.Sprintf("%0.0f - %0.0f", *minValue, *maxValue)
	case minValue != nil:
		return fmt.Sprintf("%0.0f - ∞", *minValue)
	case maxValue != nil:
		return fmt.Sprintf("0 - %0.0f", *maxValue)
	}

	return anyValue
}
