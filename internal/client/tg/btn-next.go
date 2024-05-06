package tg

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

const (
	nextPage = "next_page"
)

func (s *service) nextSettingPageBtn(c tele.Context) error {
	filter, err := s.service.ActiveFilter(s.ctx, userFromContext(c))
	if err != nil {
		return err
	}

	return s.sendSettingFilter(c, filter)
}

func nextInlineBtn(currentPageIdx, numberOfPages int, pageType string) tele.Btn {
	if numberOfPages == 1 {
		return tele.Btn{}
	}
	nextPageIdx := nextPageIdx(currentPageIdx, numberOfPages)

	return tele.Btn{
		Text: fmt.Sprintf("⚙️ (%v/%v)", currentPageIdx+1, numberOfPages),
		Data: actionData(pageType, nextPage, strconv.Itoa(nextPageIdx)),
	}
}
