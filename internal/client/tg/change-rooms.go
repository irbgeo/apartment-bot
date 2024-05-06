package tg // nolint: dupl

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	changeMinRooms = "change_min_rooms"
	changeMaxRooms = "change_max_rooms"
)

func (s *service) changeRoomsInit(isMinRooms bool) initFunc {
	return func(c tele.Context) error {
		actionType := changeMinRooms
		if !isMinRooms {
			actionType = changeMaxRooms
		}

		userID := c.Sender().ID
		s.userAction.Store(userID, actionType)

		messageText := "Enter new min rooms"
		if !isMinRooms {
			messageText = "Enter new max rooms"
		}

		msg := &tele.Message{
			Sender:      c.Sender(),
			Text:        messageText,
			ReplyMarkup: cancelOrResetMarkup(actionType),
		}

		return s.sendMessage(msg, actionMessage)
	}
}

func (s *service) changeRooms(isMinRooms bool) changeFunc {
	return func(c tele.Context) error {
		r := &client.ChangeFilterRoomsInfo{
			User:        userFromContext(c),
			IsMinChange: isMinRooms,
		}

		values := getValue(c)
		if len(values) == 0 || values[0] != anyValue {
			rooms, err := strconv.ParseFloat(c.Text(), 64)
			if err != nil {
				return fmt.Errorf("invalid value: %s", c.Text())
			}

			if r.IsMinChange {
				r.NewMinRooms = &rooms
			} else {
				r.NewMaxRooms = &rooms
			}
		}

		filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterRooms)
		if err != nil {
			return err
		}

		s.userAction.Delete(c.Sender().ID)

		return s.sendSettingFilter(c, filter)
	}
}

func (s *service) changeRoomsBtn(isMinRooms bool) func(f *server.Filter) tele.Btn {
	return func(_ *server.Filter) tele.Btn {
		text := "Min rooms"
		data := changeMinRooms
		if !isMinRooms {
			text = "Max rooms"
			data = changeMaxRooms
		}

		return tele.Btn{
			Text: text,
			Data: data,
		}
	}
}

func (s *service) roomsParamToString(f *server.Filter) string {
	param := []string{"Rooms: ", rangeStr(f.MinRooms, f.MaxRooms)}
	return strings.Join(param, "")
}
