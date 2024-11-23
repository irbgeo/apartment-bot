package client

import (
	"sync"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

type SaveFilterInfo struct {
	User *server.User
}

type channels struct {
	apartment chan server.Apartment
}

type storage struct {
	active            sync.Map
	turnedOff         turnedOffStorage
	historyReceiving  sync.Map
	disconnectedUsers sync.Map
	cities            citiesStorage
}

type turnedOffStorage struct {
	sync.RWMutex
	filters map[int64]map[string]time.Time
}

type citiesStorage struct {
	sync.Map
	firstCities []string
}
