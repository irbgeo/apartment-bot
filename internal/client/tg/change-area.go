package tg

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	changeMinArea = "change_min_area"
	changeMaxArea = "change_max_area"
)

func (s *service) changeAreaInit(isMinArea bool) initFunc {
	return func(c tele.Context) error {
		actionType := changeMinArea
		if !isMinArea {
			actionType = changeMaxArea
		}

		userID := c.Sender().ID
		s.userAction.Store(userID, actionType)

		messageText := "Enter new min area of your feature apartment (m2)"
		if !isMinArea {
			messageText = "Enter new max area of your feature apartment (m2)"
		}

		msg := &tele.Message{
			Sender:      c.Sender(),
			Text:        messageText,
			ReplyMarkup: cancelOrResetMarkup(actionType),
		}

		return s.sendMessage(msg, actionMessage)
	}
}

func (s *service) changeArea(isMinArea bool) changeFunc {
	return func(c tele.Context) error {
		r := &client.ChangeFilterAreaInfo{
			User:        userFromContext(c),
			IsMinChange: isMinArea,
		}

		values := getValue(c)
		if len(values) == 0 || values[0] != anyValue {
			area, err := strconv.ParseFloat(c.Text(), 64)
			if err != nil {
				return fmt.Errorf("invalid value: %s", c.Text())
			}

			if r.IsMinChange {
				r.NewMinArea = &area
			} else {
				r.NewMaxArea = &area
			}
		}

		filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterArea)
		if err != nil {
			return err
		}

		s.userAction.Delete(c.Sender().ID)

		return s.sendSettingFilter(c, filter)
	}
}

func (s *service) changeAreaBtn(isMinArea bool) func(f *server.Filter) tele.Btn {
	return func(_ *server.Filter) tele.Btn {
		text := "🏠 Min Area(m2)"
		data := changeMinArea
		if !isMinArea {
			text = "🏠 Max Area(m2)"
			data = changeMaxArea
		}

		return tele.Btn{
			Text: text,
			Data: data,
		}
	}
}

func (s *service) areaParamToString(f *server.Filter) string {
	param := []string{"Area: ", rangeStr(f.MinArea, f.MaxArea), " m²"}
	return strings.Join(param, "")
}
