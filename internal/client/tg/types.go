package tg

import (
	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

type MessageType int64

const (
	settingFilterMessage MessageType = iota + 1
	actionMessage
	botMessage
	userMassage
	errMessage
)

type param struct {
	init     initFunc
	change   changeFunc
	toString paramStrFunc
}

type initFunc func(c tele.Context) error
type changeFunc func(c tele.Context) error
type paramStrFunc func(f *server.Filter) string

type Message struct {
	UserID int64
	Text   string
}
