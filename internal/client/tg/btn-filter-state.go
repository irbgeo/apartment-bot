package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeStateBtn = "change_state_btn"
)

func (s *service) changeStateBtn(c tele.Context) error {
	r := &client.ChangeStateFilterInfo{
		User: userFromContext(c),
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeStateFilter)
	if err != nil {
		return err
	}

	return s.sendSettingFilter(c, filter)
}

func stateInlineBtn(f *server.Filter) tele.Btn {
	isPlay := f.PauseTimestamp == nil

	emoji := "▶️"
	if isPlay {
		emoji = "⏸️"
	}

	return tele.Btn{
		Text: emoji,
		Data: changeStateBtn,
	}
}
