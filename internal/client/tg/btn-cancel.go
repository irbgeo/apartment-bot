package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	btnCancel = "btn_cancel"
)

func (s *service) cancelBtn(c tele.Context) error {
	userID := c.Sender().ID
	_, isExist := s.userAction.Load(userID)
	if isExist {
		if err := s.cleanUserActions(userID); err != nil {
			return err
		}
		return nil
	}

	s.service.CancelCreatingFilter(s.ctx, &server.User{ID: userID})

	if err := s.cleanUserMessages(userID); err != nil {
		return err
	}

	return s.filtersListHandler(c)
}

func (s *service) cleanUserActions(userID int64) error {
	if err := s.messages.CleanTill(userID, settingFilterMessage); err != nil {
		return err
	}
	s.userAction.Delete(userID)
	return nil
}

func cancelInlineBtn() tele.Btn {
	return tele.Btn{
		Text: "🚫",
		Data: btnCancel,
	}
}
