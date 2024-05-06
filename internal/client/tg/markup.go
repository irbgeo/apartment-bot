package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	filterSetting = "filter_setting"
)

func (s *service) filterSettingMarkup(f *server.Filter, settingPageIdx int) *tele.ReplyMarkup {
	filterMarkup := &tele.ReplyMarkup{
		ForceReply: true,
	}
	settingRows := s.settingsRows(f, settingPageIdx)
	controlRow := s.controlRow(f, settingPageIdx)
	filterMarkup.Inline(append(settingRows, controlRow)...)
	return filterMarkup
}

func (s *service) settingsRows(f *server.Filter, settingPageIdx int) []tele.Row {
	settings := make([]tele.Row, 0, len(s.settingBtns[settingPageIdx]))

	for _, rowBtn := range s.settingBtns[settingPageIdx] {
		settingRow := make(tele.Row, 0, len(rowBtn))
		for _, btn := range rowBtn {
			settingRow = append(settingRow, btn(f))
		}
		settings = append(settings, settingRow)
	}

	return settings
}

func (s *service) controlRow(f *server.Filter, settingPageIdx int) tele.Row {
	controlRow := tele.Row{cancelInlineBtn()}

	if f.ID != "" {
		deleteBtn := deleteInlineBtn(f.ID)
		controlRow = append(controlRow, deleteBtn)
	}

	stateBtn := stateInlineBtn(f)
	nextPageBtn := nextInlineBtn(settingPageIdx, len(s.settingBtns), filterSetting)
	controlRow = append(controlRow, stateBtn, nextPageBtn)

	if f.IsUpdate {
		okBtn := okInlineBtn()
		controlRow = append(controlRow, okBtn)
	}

	return controlRow
}

func (s *service) filterMenu(filters []server.Filter) *tele.ReplyMarkup {
	rows := make([]tele.Row, 0, len(filters))

	for _, f := range filters {
		rows = append(rows, tele.Row{tele.Btn{Text: *f.Name}})
	}

	m := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		RemoveKeyboard: true,
		ForceReply:     true,
	}
	m.Reply(rows...)

	return m
}

func filterSavedMarkup(filterID string, count int64) *tele.ReplyMarkup {
	rows := make([]tele.Row, 0)

	if count > 0 {
		rows = append(rows, tele.Row{getOldApartmentsInlineBtn(filterID, count)})
	}

	rows = append(rows, tele.Row{getNewApartmentsInlineBtn()})

	m := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		RemoveKeyboard: true,
		ForceReply:     true,
	}
	m.Inline(rows...)

	return m
}

func cancelOrResetMarkup(actionType string) *tele.ReplyMarkup {
	m := &tele.ReplyMarkup{}
	m.Inline(
		tele.Row{cancelInlineBtn(), resetInlineBtn(actionType)},
	)
	return m
}
