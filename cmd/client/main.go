package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	apiserver "github.com/irbgeo/apartment-bot/internal/api/server"
	"github.com/irbgeo/apartment-bot/internal/client"
	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
	"github.com/irbgeo/apartment-bot/internal/client/tg/message"
)

type configuration struct {
	ServerURL                                string        `envconfig:"SERVER_URL" default:"localhost:9000"`
	MessageURL                               string        `envconfig:"MESSAGE_URL" default:"localhost:9001"`
	TelegramBotSecret                        string        `envconfig:"TELEGRAM_BOT_SECRET" required:"true"`
	TelegramBotSendPeriod                    time.Duration `envconfig:"TELEGRAM_BOT_SEND_PERIOD" default:"10s"`
	TelegramBotMaxCountSendMessagesPerPeriod int           `envconfig:"TELEGRAM_BOT_MAX_COUNT_SEND_MESSAGES_PER_PERIOD" default:"3"`
	TelegramBotAdminUsername                 string        `envconfig:"TELEGRAM_BOT_ADMIN_USERNAME" default:"rent_apartment_georgia_bot_admin"`
	TelegramBotDisabledParameters            []string      `envconfig:"TELEGRAM_BOT_DISABLED_PARAMS" default:""`
	FirstCities                              []string      `envconfig:"FIRST_CITIES" default:"Tbilisi,Batumi"`
	AuthToken                                string        `envconfig:"AUTH_TOKEN" default:"LdBD1e0q"`
	ClientTag                                int64         `envconfig:"CLIENT_TAG" default:"1"`
}

func main() {
	slog.Info("Hi!")

	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("read configuration", "err", err)
		os.Exit(1)
	}

	slog.Info("configuration", "cfg", cfg)

	serverCli, err := apiserver.NewClient(cfg.ServerURL, cfg.AuthToken, cfg.ClientTag)
	if err != nil {
		slog.Error("init server cli", "err", err)
		os.Exit(1)
	}

	cli, err := client.NewService(serverCli, cfg.FirstCities)
	if err != nil {
		slog.Error("init client", "err", err)
		os.Exit(1)
	}

	if err := cli.Start(); err != nil {
		slog.Error("start client", "err", err)
		os.Exit(1)
	}
	defer cli.Stop()

	massageStack := message.NewService()

	botCfg := tgbot.StartConfig{
		Token:               cfg.TelegramBotSecret,
		DisabledParameters:  cfg.TelegramBotDisabledParameters,
		AdminUsername:       cfg.TelegramBotAdminUsername,
		MaxPhotoCount:       cfg.TelegramBotMaxCountSendMessagesPerPeriod,
		MessageSendInterval: cfg.TelegramBotSendPeriod,
	}

	b, err := tgbot.NewService(
		botCfg,
		cli,
		massageStack,
	)
	if err != nil {
		slog.Error("init bot", "err", err)
		os.Exit(1)
	}
	if err := b.Start(); err != nil {
		slog.Error("start bot", "err", err)
		os.Exit(1)
	}
	defer b.Stop()

	slog.Info("I'm turned on")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	slog.Info("Goodbye!")
}
