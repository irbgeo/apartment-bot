package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/irbgeo/apartment-bot/internal/apartment"
	"github.com/irbgeo/apartment-bot/internal/apartment/provider/ssge"
	"github.com/irbgeo/apartment-bot/internal/api/health"
	api "github.com/irbgeo/apartment-bot/internal/api/server"
	"github.com/irbgeo/apartment-bot/internal/filter"
	"github.com/irbgeo/apartment-bot/internal/server"
	"github.com/irbgeo/apartment-bot/internal/storage/mongo"
)

type configuration struct {
	Address                 string        `envconfig:"ADDRESS" default:":9000"`
	HealthAddress           string        `envconfig:"HEALTH_ADDRESS" default:":9005"`
	MongoAddress            string        `envconfig:"MONGO_ADDRESS" default:"localhost:27017"`
	MongoUsername           string        `envconfig:"MONGO_USERNAME" default:"root"`
	MongoPassword           string        `envconfig:"MONGO_PASSWORD" default:"password"`
	MongoDatabase           string        `envconfig:"MONGO_DATABASE" default:"apartment"`
	MaxFetchPages           int64         `envconfig:"MAX_FETCH_PAGES" default:"30"`
	ApartmentUpdateInterval time.Duration `envconfig:"APARTMENT_UPDATE_INTERVAL" default:"1m"`
	ApartmentDayToLive      int64         `envconfig:"APARTMENT_DAY_TO_LIVE" default:"7"`
	RefreshTokenInterval    time.Duration `envconfig:"REFRESH_TOKEN_INTERVAL" default:"10m"`
	WithRefreshApartments   bool          `envconfig:"WITH_REFRESH_APARTMENTS" default:"false"`
	AuthToken               string        `envconfig:"AUTH_TOKEN" default:"test"`
}

func main() {
	slog.Info("Hi!")

	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("read configuration", "err", err)
		os.Exit(1)
	}

	slog.Info("configuration", "cfg", cfg)

	mongoCfg := mongo.Config{
		Address:  cfg.MongoAddress,
		Username: cfg.MongoUsername,
		Password: cfg.MongoPassword,
		Database: cfg.MongoDatabase,
	}
	stor, err := mongo.NewStorage(mongoCfg)
	if err != nil {
		slog.Error("init mongo storage", "err", err)
		os.Exit(1)
	}

	filterProvider, err := filter.New(stor)
	if err != nil {
		slog.Error("init filters", "err", err)
		os.Exit(1)
	}

	ssProvider := ssge.NewSSGEProvider()

	apartmentCfg := apartment.Config{
		MaxFetchPages: cfg.MaxFetchPages,
		ApartmentTTL:  time.Duration(cfg.ApartmentDayToLive) * 24 * time.Hour,
	}

	apartmentSvc := apartment.NewService(
		apartmentCfg,
		ssProvider,
	)

	srv := server.NewService(
		apartmentSvc,
		stor,
		filterProvider,
	)

	// start provider service for refreshing access token
	if err := ssProvider.Start(cfg.RefreshTokenInterval); err != nil {
		slog.Error("start ssge provider", "err", err)
		os.Exit(1)
	}
	defer ssProvider.Stop()

	// refresh apartments
	if cfg.WithRefreshApartments {
		if err := srv.RefreshApartments(); err != nil {
			slog.Error("refresh apartments", "err", err)
			os.Exit(1)
		}
	}

	// fill in apartment service cache
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	savedApartmentCh, err := stor.Apartments(ctx, server.Filter{})
	if err != nil {
		slog.Error("get saved apartments", "err", err)
		os.Exit(1)
	}

	var cnt int
	for a := range savedApartmentCh {
		apartmentSvc.SetInCache(a)
		cnt++
	}
	slog.Info("apartments in cache", "cnt", cnt)

	// start apartment service
	if err := apartmentSvc.Start(cfg.ApartmentUpdateInterval); err != nil {
		slog.Error("start apartment service", "err", err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		slog.Error("start server", "err", err)
		os.Exit(1)
	}
	defer srv.Stop()

	// start api
	go func() {
		if err := api.ListenAndServe(cfg.Address, cfg.AuthToken, srv); err != nil {
			slog.Error("turn on server server", "err", err)
			os.Exit(1)
		}
	}()

	// start healthcheck
	go func() {
		if err := health.ListenAndServe(cfg.HealthAddress); err != nil {
			slog.Error("turn on health server", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("I'm turned on")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	slog.Info("Goodbye!")
}
