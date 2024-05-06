package tg

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeDistrict = "change_district"
)

func (s *service) changeDistrictInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeDistrict)

	f, err := s.service.ActiveFilter(s.ctx, userFromContext(c))
	if err != nil {
		return nil
	}

	values := getValue(c)
	idx := 0
	if len(values) == 2 {
		idx, _ = strconv.Atoi(values[1])
	}

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Choose district you would like to live ",
		ReplyMarkup: s.districtMarkup(f, idx),
	}

	return s.sendMessage(msg, actionMessage)
}

func (s *service) districtMarkup(f *server.Filter, pageIdx int) *tele.ReplyMarkup {
	districts := s.service.AvailableDistrictsForCity(*f.City)
	groups := group(districts)
	g := groups[pageIdx]

	rows := make([]tele.Row, 0, verticalN+1)
	row := make([]tele.Btn, 0, horizontalN)
	for _, district := range g {
		if len(row) == horizontalN {
			rows = append(rows, row)
			row = make([]tele.Btn, 0, horizontalN)
		}

		text := district
		if _, ok := f.District[district]; ok {
			text = fmt.Sprintf("✅ %v", district)
		}
		row = append(
			row,
			tele.Btn{
				Text: text,
				Data: actionData(changeDistrict, district),
			},
		)
	}
	rows = append(rows, row)

	settingRow := tele.Row{cancelInlineBtn(), resetInlineBtn(changeDistrict)}
	settingRow = append(settingRow, nextInlineBtn(pageIdx, len(groups), changeDistrict))

	rows = append(rows, settingRow)

	districtMarkup := &tele.ReplyMarkup{}
	districtMarkup.Inline(
		rows...,
	)

	return districtMarkup
}

func (s *service) changeDistrict(c tele.Context) error {
	r := &client.ChangeFilterDistrictInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) != 0 && values[0] != anyValue {
		r.ChoseDistrict = &values[0]
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterDistrict)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

var startDistrictIdx = "0"

func (s *service) changeDistrictBtn(f *server.Filter) tele.Btn {
	var text, data string

	if f.City != nil {
		districts := s.service.AvailableDistrictsForCity(*f.City)
		if len(districts) != 0 {
			text = "District"
			data = actionData(changeDistrict, startDistrictIdx)
		}
	}

	return tele.Btn{
		Text: text,
		Data: data,
	}
}

func (s *service) districtParamToString(f *server.Filter) string {
	if f.City == nil {
		return ""
	}

	param := make([]string, 0, 2+2*len(f.District))
	param = append(param, "District: ")

	if len(f.District) == 0 {
		param = append(param, anyValue)
		return strings.Join(param, "")
	}

	for district := range f.District {
		param = append(param, "✅ "+district)
	}

	return strings.Join(param, "\n")
}
