package tg

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	changeMinPrice = "change_min_price"
	changeMaxPrice = "change_max_price"
)

func (s *service) changePriceInit(isMinPrice bool) initFunc {
	return func(c tele.Context) error {
		actionType := changeMinPrice
		if !isMinPrice {
			actionType = changeMaxPrice
		}

		userID := c.Sender().ID
		s.userAction.Store(userID, actionType)

		messageText := "Enter new min price"
		if !isMinPrice {
			messageText = "Enter new max price"
		}

		msg := &tele.Message{
			Sender:      c.Sender(),
			Text:        messageText,
			ReplyMarkup: cancelOrResetMarkup(actionType),
		}

		return s.sendMessage(msg, actionMessage)
	}
}

func (s *service) changePrice(isMinPrice bool) changeFunc {
	return func(c tele.Context) error {
		r := &client.ChangeFilterPriceInfo{
			User:        userFromContext(c),
			IsMinChange: isMinPrice,
		}

		values := getValue(c)
		if len(values) == 0 || values[0] != anyValue {
			price, err := strconv.ParseFloat(c.Text(), 64)
			if err != nil {
				return fmt.Errorf("invalid value: %s", c.Text())
			}

			if r.IsMinChange {
				r.NewMinPrice = &price
			} else {
				r.NewMaxPrice = &price
			}
		}

		filter, err := client.WithActiveFilter(s.ctx, r, s.service.ChangeFilterPrice)
		if err != nil {
			return err
		}

		s.userAction.Delete(c.Sender().ID)

		return s.sendSettingFilter(c, filter)
	}
}

func (s *service) changePriceBtn(isMinPrice bool) func(f *server.Filter) tele.Btn {
	return func(_ *server.Filter) tele.Btn {
		text := "💲 Min price"
		data := changeMinPrice
		if !isMinPrice {
			text = "💲 Max price"
			data = changeMaxPrice
		}

		return tele.Btn{
			Text: text,
			Data: data,
		}
	}
}

func (s *service) priceParamToString(f *server.Filter) string {
	param := []string{"Price: ", rangeStr(f.MinPrice, f.MaxPrice), " $"}
	return strings.Join(param, "")
}
