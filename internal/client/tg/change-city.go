package tg

import (
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeCity = "change_city"
)

func (s *service) changeCityInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeCity)

	cities := s.service.AvailableCities()

	values := getValue(c)
	idx := 0
	if len(values) == 2 {
		idx, _ = strconv.Atoi(values[1])
	}

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Choose city you would like to live ",
		ReplyMarkup: cityMarkup(cities, idx),
	}

	return s.sendMessage(msg, actionMessage)
}

func cityMarkup(cities []string, pageIdx int) *tele.ReplyMarkup {
	groups := group(cities)
	g := groups[pageIdx]

	rows := make([]tele.Row, 0, verticalN+1)
	row := make([]tele.Btn, 0, horizontalN)
	for _, city := range g {
		if len(row) == horizontalN {
			rows = append(rows, row)
			row = make([]tele.Btn, 0, horizontalN)
		}
		row = append(
			row,
			tele.Btn{
				Text: city,
				Data: actionData(changeCity, city),
			},
		)
	}
	rows = append(rows, row)

	settingRow := tele.Row{
		cancelInlineBtn(),
		resetInlineBtn(changeCity),
		nextInlineBtn(pageIdx, len(groups), changeCity),
	}

	rows = append(rows, settingRow)

	cityMarkup := &tele.ReplyMarkup{}
	cityMarkup.Inline(
		rows...,
	)

	return cityMarkup
}

func (s *service) changeCity(c tele.Context) error {
	r := &client.ChangeFilterCityInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) != 0 && values[0] != anyValue {
		r.NewCity = &values[0]
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterCity)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

var startCityIdx = "0"

func changeCityBtn(_ *server.Filter) tele.Btn {
	return tele.Btn{
		Text: "🏙️ City",
		Data: actionData(changeCity, startCityIdx),
	}
}

func (s *service) cityParamToString(f *server.Filter) string {
	param := make([]string, 0, 2)
	param = append(param, "City: ")

	if f.City == nil {
		param = append(param, anyValue)
	} else {
		param = append(param, *f.City)
	}

	return strings.Join(param, "")
}
