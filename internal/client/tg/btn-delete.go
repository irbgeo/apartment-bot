package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	btnDelete = "btn_delete"
)

func (s *service) deleteBtn(c tele.Context) error {
	value := getValue(c)[0]
	userID := c.Sender().ID
	f := &server.Filter{
		ID: value,
		User: &server.User{
			ID: userID,
		},
	}

	if err := s.service.DeleteFilter(s.ctx, f); err != nil {
		return err
	}

	if err := s.cleanUserMessages(userID); err != nil {
		return err
	}

	return s.filtersListHandler(c)
}

func deleteInlineBtn(value string) tele.Btn {
	return tele.Btn{
		Text: "🗑️",
		Data: actionData(btnDelete, value),
	}
}
