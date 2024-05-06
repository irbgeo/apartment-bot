package tg

import (
	tele "gopkg.in/telebot.v3"
)

const (
	btnGetNewApartments = "btn_get_new"
)

func (s *service) getNewApartmentsBtn(c tele.Context) error {
	userID := c.Sender().ID

	if err := s.cleanUserMessages(userID); err != nil {
		return err
	}
	return s.filtersListHandler(c)
}

func getNewApartmentsInlineBtn() tele.Btn {
	return tele.Btn{
		Text: "Getting only new apartments",
		Data: btnGetNewApartments,
	}
}
