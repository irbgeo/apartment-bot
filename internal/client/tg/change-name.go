package tg

import (
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeName = "change_name"
)

func (s *service) changeNameInit(c tele.Context) error {
	userID := c.Sender().ID
	s.userAction.Store(userID, changeName)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Enter new name for your filter",
		ReplyMarkup: cancelOrResetMarkup(changeName),
	}
	return s.sendMessage(msg, actionMessage)
}

func (s *service) changeName(c tele.Context) error {
	r := &client.ChangeFilterNameInfo{
		User: userFromContext(c),
	}

	name := c.Text()
	if name != anyValue {
		r.NewName = &name
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterName)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func changeNameBtn(_ *server.Filter) tele.Btn {
	return tele.Btn{
		Text: "Filter name",
		Data: changeName,
	}
}

func (s *service) nameParamToString(f *server.Filter) string {
	param := make([]string, 0, 3)

	state := "⏸️"
	if f.PauseTimestamp == nil {
		state = "▶️"
	}
	param = append(param, state)

	param = append(param, " #")

	if f.Name == nil {
		param = append(param, unknownValue)
	} else {
		param = append(param, *f.Name)
	}

	return strings.Join(param, "")
}
