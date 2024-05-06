package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	api "github.com/irbgeo/apartment-bot/internal/api/message"
	"github.com/irbgeo/apartment-bot/internal/message"
	"github.com/irbgeo/apartment-bot/internal/storage/mongo"
)

type configuration struct {
	Address   string `envconfig:"ADDRESS" default:":9001"`
	AuthToken string `envconfig:"AUTH_TOKEN" default:"test"`
	MongoURL  string `envconfig:"MONGO_URL" default:"mongodb://apartment:apartment@localhost:27017"`
	MongoDB   string `envconfig:"MONGO_DB" default:"apartment"`
}

func main() {
	slog.Info("Hi!")

	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("read configuration", "err", err)
		os.Exit(1)
	}

	slog.Info("configuration", "cfg", cfg)

	stor, err := mongo.NewStorage(cfg.MongoURL, cfg.MongoDB)
	if err != nil {
		slog.Error("init mongo storage", "err", err)
		os.Exit(1)
	}

	msgSvc := message.NewService(stor)
	defer msgSvc.Close()

	go func() {
		if err := api.ListenAndServe(cfg.Address, cfg.AuthToken, msgSvc); err != nil {
			slog.Error("turn on server server", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("I'm turned on")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	slog.Info("Goodbye!")
}
