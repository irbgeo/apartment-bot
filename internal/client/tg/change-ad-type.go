package tg

import (
	"strconv"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeAdType = "change_ad_type"

	adTypeString = map[int64]string{
		server.RentAdType: "For Rent",
		server.SaleAdType: "For Sale",
	}
)

func (s *service) changeTypeInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeAdType)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Choose what advertisement are you looking for ",
		ReplyMarkup: typeMarkup(),
	}

	return s.sendMessage(msg, actionMessage)
}

func typeMarkup() *tele.ReplyMarkup {
	rows := []tele.Row{
		{
			{
				Text: "For Rent",
				Data: actionData(changeAdType, strconv.FormatInt(server.RentAdType, 10)),
			},
		},
		{
			{
				Text: "For Sale",
				Data: actionData(changeAdType, strconv.FormatInt(server.SaleAdType, 10)),
			},
		},
	}

	rows = append(rows, tele.Row{cancelInlineBtn(), resetInlineBtn(changeAdType)})

	typeMarkup := &tele.ReplyMarkup{}
	typeMarkup.Inline(
		rows...,
	)

	return typeMarkup
}

func (s *service) changeAdType(c tele.Context) error {
	r := &client.ChangeAdTypeFilterInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) != 0 {
		t, _ := strconv.ParseInt(values[0], 10, 64)
		r.NewAdType = &t
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeTypeFilter)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func (s *service) changeAdTypeBtn(_ *server.Filter) tele.Btn {
	if _, isExist := s.params[changeAdType]; isExist {
		return tele.Btn{
			Text: "Advertisement Type",
			Data: changeAdType,
		}
	}

	return tele.Btn{}
}

func (s *service) adTypeParamToString(f *server.Filter) string {
	param := "Advertisement Type: "

	if f.AdType == nil {
		if _, ok := s.params[changeAdType]; !ok {
			return ""
		}
		return param + anyValue
	}

	return param + adTypeString[*f.AdType]
}
