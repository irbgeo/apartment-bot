package tg

import (
	"fmt"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	btnGetOldApartments = "btn_get_old"
)

func (s *service) getOldApartmentsBtn(c tele.Context) error {
	userID := c.Sender().ID
	f := &server.Filter{
		ID: getValue(c)[0],
		User: &server.User{
			ID: userID,
		},
	}

	err := s.service.Apartments(s.ctx, f)
	if err != nil {
		return err
	}

	err = s.messages.CleanUserMessages(userID)
	if err != nil {
		return err
	}
	return s.filtersListHandler(c)
}

func getOldApartmentsInlineBtn(filterID string, count int64) tele.Btn {
	return tele.Btn{
		Text: fmt.Sprintf("Get previous apartments (%d)", count),
		Data: actionData(btnGetOldApartments, filterID),
	}
}
