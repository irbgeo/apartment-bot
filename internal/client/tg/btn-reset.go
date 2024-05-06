package tg

import tele "gopkg.in/telebot.v3"

func resetInlineBtn(actionType string) tele.Btn {
	return tele.Btn{
		Text: "reset",
		Data: actionData(actionType, anyValue),
	}
}
