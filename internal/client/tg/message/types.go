package message

import (
	tele "gopkg.in/telebot.v3"

	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
)

// message represents a telegram message with its type.
type message struct {
	Content *tele.Message
	Type    tgbot.MessageType
}
