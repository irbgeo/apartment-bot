package tg

import (
	"strconv"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	changeBuildingStatus = "change_building_status"

	buildingStatusString = map[int64]string{
		server.NewBuildingStatus:               "New",
		server.UnderConstructionBuildingStatus: "Under Construction",
		server.OldBuildingStatus:               "Old",
	}
)

func (s *service) changeBuildingStatusInit(c tele.Context) error {
	userID := c.Sender().ID

	s.userAction.Store(userID, changeBuildingStatus)

	msg := &tele.Message{
		Sender:      c.Sender(),
		Text:        "Choose what building you are looking for",
		ReplyMarkup: statusMarkup(),
	}

	return s.sendMessage(msg, actionMessage)
}

func statusMarkup() *tele.ReplyMarkup {
	rows := []tele.Row{
		{
			{
				Text: "New",
				Data: actionData(changeBuildingStatus, strconv.FormatInt(server.NewBuildingStatus, 10)),
			},
		},
		{
			{
				Text: "Under constructions",
				Data: actionData(changeBuildingStatus, strconv.FormatInt(server.UnderConstructionBuildingStatus, 10)),
			},
		},
		{
			{
				Text: "Old",
				Data: actionData(changeBuildingStatus, strconv.FormatInt(server.OldBuildingStatus, 10)),
			},
		},
	}

	rows = append(rows, tele.Row{cancelInlineBtn(), resetInlineBtn(changeBuildingStatus)})

	statusMarkup := &tele.ReplyMarkup{}
	statusMarkup.Inline(rows...)

	return statusMarkup
}

func (s *service) changeBuildingStatus(c tele.Context) error {
	r := &client.ChangeBuildingStatusFilterInfo{
		User: userFromContext(c),
	}

	values := getValue(c)
	if len(values) != 0 && values[0] != anyValue {
		t, _ := strconv.ParseInt(values[0], 10, 64)
		r.NewBuildingStatus = &t
	}

	filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeBuildingStatusFilter)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func (s *service) changeBuildingStatusBtn(_ *server.Filter) tele.Btn {
	return tele.Btn{
		Text: "Building Status",
		Data: changeBuildingStatus,
	}
}

func (s *service) buildingStatusParamToString(f *server.Filter) string {
	param := "Building Status: "

	if f.BuildingStatus == nil {
		return param + anyValue
	}

	status, ok := buildingStatusString[*f.BuildingStatus]
	if !ok {
		return ""
	}

	return param + status
}
