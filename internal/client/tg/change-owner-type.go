package tg

import (
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeOwnerType = "change_owner_type"
)

func (s *service) changeOwnerTypeInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeOwnerType)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Choose who you would like to receive ads from",
		ReplyMarkup: ownerTypeMarkup(),
	}

	return s.sendMessage(msg, actionMessage)
}

func ownerTypeMarkup() *tele.ReplyMarkup {
	rows := make([]tele.Row, 0, len(ownerTypeMap)+2)

	for _, ownerType := range ownerTypeMap {
		rows = append(
			rows,
			tele.Row{
				tele.Btn{
					Text: ownerType,
					Data: actionData(changeOwnerType, ownerType),
				},
			},
		)
	}

	rows = append(rows, tele.Row{cancelInlineBtn(), resetInlineBtn(changeOwnerType)})

	ownerTypeMarkup := &tele.ReplyMarkup{}

	ownerTypeMarkup.Inline(
		rows...,
	)

	return ownerTypeMarkup
}

func (s *service) changeOwnerType(c tele.Context) error {
	r := &client.ChangeOwnerTypeFilterInfo{
		User: userFromContext(c),
	}

	values := getValue(c)

	if len(values) != 0 && values[0] != anyValue {
		ownerType := values[0] == "Owner"
		r.NewOwnerType = &ownerType
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeOwnerTypeFilter)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func (s *service) changeOwnerTypeBtn(_ *server.Filter) tele.Btn {
	if _, isExist := s.params[changeAdType]; isExist {
		return tele.Btn{
			Text: "Owner",
			Data: changeOwnerType,
		}
	}

	return tele.Btn{}
}

var ownerTypeString = map[bool]string{
	true:  "Owner",
	false: "Agency",
}

func (s *service) ownerTypeParamToString(f *server.Filter) string {
	if _, ok := s.params[changeOwnerType]; !ok {
		return ""
	}

	param := make([]string, 0, 2)
	param = append(param, "Owner: ")

	if f.IsOwner == nil {
		param = append(param, anyValue)
	} else {
		param = append(param, ownerTypeString[*f.IsOwner])
	}

	return strings.Join(param, "")
}
