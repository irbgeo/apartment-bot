package message

import (
	api "github.com/irbgeo/apartment-bot/internal/api/message/proto"
	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
	"github.com/irbgeo/apartment-bot/internal/message"
)

func messageFromAPIToBot(in *api.Message) tgbot.Message {
	return tgbot.Message{
		UserID: in.UserId,
		Text:   in.Text,
	}
}

func messageFromBotToSvc(in *api.Message) message.Message {
	return message.Message{
		UserID: in.UserId,
		Text:   in.Text,
	}
}

func messageFromSvcToBot(in message.Message) *api.Message {
	return &api.Message{
		UserId: in.UserID,
		Text:   in.Text,
	}
}
