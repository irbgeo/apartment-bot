package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
)

const (
	btnOk = "btn_ok"
)

func (s *service) okBtn(c tele.Context) error {
	r := &client.SaveFilterInfo{
		User: userFromContext(c),
	}

	filter, count, err := s.service.SaveFilter(s.ctx, r)
	if err == client.ErrUnknownFilterName {
		if err := s.handleUnknownFilterName(c); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	if err := s.sendSavedFilter(c, count, filter); err != nil {
		return err
	}

	return nil
}

func (s *service) handleUnknownFilterName(c tele.Context) error {
	err := s.changeNameInit(c)
	if err != nil {
		return err
	}

	return nil
}

func okInlineBtn() tele.Btn {
	return tele.Btn{
		Text: "✅",
		Data: btnOk,
	}
}
