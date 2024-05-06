package tg

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeLocation    = "change_location"
	changeMaxDistance = "change_max_distance"
)

func (s *service) changeLocationInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeLocation)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Send the location of place you would like to live nearby\n⚠️ Send location from Telegram",
		ReplyMarkup: cancelOrResetMarkup(changeLocation),
	}

	return s.sendMessage(msg, actionMessage)
}

func (s *service) changeLocation(c tele.Context) error {
	r := &client.ChangeFilterLocationInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) == 0 || values[0] != anyValue {
		l := c.Message().Location
		if l == nil {
			return errNotFoundLocation
		}

		r.NewCoordinates = &client.Coordinates{
			Lat: float64(l.Lat),
			Lng: float64(l.Lng),
		}
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterLocation)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func changeLocationBtn(_ *server.Filter) tele.Btn {
	return tele.Btn{
		Text: "📍 Location",
		Data: changeLocation,
	}
}

const locationMapPrefix = "https://www.google.com/maps/search/?api=1&query="

func (s *service) locationParamToString(f *server.Filter) string {
	params := make([]string, 0, 3)
	params = append(params, "Location: ")

	if f.Coordinates == nil {
		params = append(params, "Not set")
	} else {
		params = append(params, locationString(f.Coordinates.Lat, f.Coordinates.Lng))
	}

	return strings.Join(params, "")
}

func (s *service) changeMaxDistanceInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeMaxDistance)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Enter the maximum distance to the location you would like to live nearby (m)",
		ReplyMarkup: cancelOrResetMarkup(changeMaxDistance),
	}

	return s.sendMessage(msg, actionMessage)
}

func (s *service) changeMaxDistance(c tele.Context) error {
	r := &client.ChangeFilterMaxDistanceInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) == 0 || values[0] != anyValue {
		minDistance, err := strconv.ParseFloat(c.Text(), 64)
		if err != nil {
			return fmt.Errorf("invalid value: %s", c.Text())
		}

		r.NewMaxDistance = &minDistance
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterMaxDistance)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func changeMaxDistanceBtn(f *server.Filter) tele.Btn {
	var btn tele.Btn

	if f.Coordinates != nil {
		btn = tele.Btn{
			Text: "Max distance",
			Data: changeMaxDistance,
		}
	}
	return btn
}

func (s *service) maxDistanceParamToString(f *server.Filter) string {
	param := make([]string, 0, 2)
	param = append(param, "Max distance: ")

	if f.MaxDistance == nil {
		param = append(param, anyValue)
	} else {
		param = append(param, strconv.FormatFloat(*f.MaxDistance, 'f', -1, 64), " m")
	}

	return strings.Join(param, "")
}
