package tg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/client"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	botTimeout = 10 * time.Second
)

type service struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	b                   *tele.Bot
	service             apartmentSvc
	adminUsername       string
	maxPhotoCount       int
	messageSendInterval time.Duration
	userAction          sync.Map
	messages            messageStack
	params              map[string]param
	btn                 map[string]changeFunc
	settingBtns         [][][]func(f *server.Filter) tele.Btn
	sendMessageCh       chan Message
}

//go:generate mockery --name apartmentSvc --structname ApartmentSvc
type apartmentSvc interface {
	Watcher() <-chan server.Apartment
	StartChat(ctx context.Context, u *server.User) error
	Filters(ctx context.Context, u *server.User) ([]server.Filter, error)
	ActiveFilter(ctx context.Context, u *server.User) (*server.Filter, error)
	Filter(ctx context.Context, f *server.Filter) (*server.Filter, error)
	StartCreatingFilter(ctx context.Context, u *server.User) *server.Filter
	ChangeFilterName(ctx context.Context, i *client.ChangeFilterNameInfo) (*server.Filter, error)
	ChangeBuildingStatusFilter(ctx context.Context, i *client.ChangeBuildingStatusFilterInfo) (*server.Filter, error)
	ChangeTypeFilter(ctx context.Context, i *client.ChangeAdTypeFilterInfo) (*server.Filter, error)
	ChangeFilterCity(ctx context.Context, i *client.ChangeFilterCityInfo) (*server.Filter, error)
	ChangeFilterDistrict(ctx context.Context, i *client.ChangeFilterDistrictInfo) (*server.Filter, error)
	ChangeFilterPrice(ctx context.Context, i *client.ChangeFilterPriceInfo) (*server.Filter, error)
	ChangeFilterRooms(ctx context.Context, i *client.ChangeFilterRoomsInfo) (*server.Filter, error)
	ChangeFilterArea(ctx context.Context, i *client.ChangeFilterAreaInfo) (*server.Filter, error)
	ChangeFilterLocation(ctx context.Context, i *client.ChangeFilterLocationInfo) (*server.Filter, error)
	ChangeFilterMaxDistance(ctx context.Context, i *client.ChangeFilterMaxDistanceInfo) (*server.Filter, error)
	ChangeStateFilter(ctx context.Context, i *client.ChangeStateFilterInfo) (*server.Filter, error)
	ChangeOwnerTypeFilter(ctx context.Context, i *client.ChangeOwnerTypeFilterInfo) (*server.Filter, error)
	CancelCreatingFilter(ctx context.Context, u *server.User)
	SaveFilter(ctx context.Context, i *client.SaveFilterInfo) (*server.Filter, int64, error)
	DeleteFilter(ctx context.Context, f *server.Filter) error
	Apartments(ctx context.Context, f *server.Filter) error
	AvailableCities() []string
	AvailableDistrictsForCity(city string) []string
	WorkingFilters(userID int64, f []string) []string

	IsAllow(userID int64) bool

	ErrorHandler(ctx context.Context, u *server.User, err error)
	FloodErrorHandler(ctx context.Context, u *server.User, retryAt time.Duration)
	BlockErrorHandler(ctx context.Context, u *server.User, err error)
}

//go:generate mockery --name messageStack --structname MessageStack
type messageStack interface {
	SetBot(b *tele.Bot)
	StoreMessage(userID int64, m *tele.Message, t MessageType)
	GetOrCleanTill(userID int64, goalType, tillType MessageType) (*tele.Message, bool, error)
	CleanUserMessages(userID int64) error
	CleanMessagesUntil(userID int64, t MessageType) error
}

func NewService(
	cfg StartConfig,
	aSvc apartmentSvc,
	mStack messageStack,
) (*service, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  cfg.Token,
		Poller: &tele.LongPoller{Timeout: botTimeout},
	})
	if err != nil {
		return nil, err
	}

	mStack.SetBot(b)

	t := &service{
		b:                   b,
		messages:            mStack,
		service:             aSvc,
		adminUsername:       cfg.AdminUsername,
		maxPhotoCount:       cfg.MaxPhotoCount,
		messageSendInterval: cfg.MessageSendInterval,
		sendMessageCh:       make(chan Message),
	}

	t.initParams(cfg.DisabledParameters)
	t.initBtns()
	t.initHandlers()

	return t, nil
}

func (s *service) initParams(disabledParameters []string) {
	s.params = map[string]param{
		changeName: {
			init:     s.changeNameInit,
			change:   s.changeName,
			toString: s.nameParamToString,
		},
		changeAdType: {
			init:     s.changeTypeInit,
			change:   s.changeAdType,
			toString: s.adTypeParamToString,
		},
		changeBuildingStatus: {
			init:     s.changeBuildingStatusInit,
			change:   s.changeBuildingStatus,
			toString: s.buildingStatusParamToString,
		},
		changeCity: {
			init:     s.changeCityInit,
			change:   s.changeCity,
			toString: s.cityParamToString,
		},
		changeDistrict: {
			init:     s.changeDistrictInit,
			change:   s.changeDistrict,
			toString: s.districtParamToString,
		},
		changeMinPrice: {
			init:     s.changePriceInit(true),
			change:   s.changePrice(true),
			toString: s.priceParamToString,
		},
		changeMaxPrice: {
			init:   s.changePriceInit(false),
			change: s.changePrice(false),
		},
		changeMinRooms: {
			init:     s.changeRoomsInit(true),
			change:   s.changeRooms(true),
			toString: s.roomsParamToString,
		},
		changeMaxRooms: {
			init:   s.changeRoomsInit(false),
			change: s.changeRooms(false),
		},
		changeMinArea: {
			init:     s.changeAreaInit(true),
			change:   s.changeArea(true),
			toString: s.areaParamToString,
		},
		changeMaxArea: {
			init:   s.changeAreaInit(false),
			change: s.changeArea(false),
		},
		changeLocation: {
			init:     s.changeLocationInit,
			change:   s.changeLocation,
			toString: s.locationParamToString,
		},
		changeMaxDistance: {
			init:     s.changeMaxDistanceInit,
			change:   s.changeMaxDistance,
			toString: s.maxDistanceParamToString,
		},
		changeOwnerType: {
			init:     s.changeOwnerTypeInit,
			change:   s.changeOwnerType,
			toString: s.ownerTypeParamToString,
		},
	}

	for _, param := range disabledParameters {
		delete(s.params, param)
	}
}

func (s *service) initBtns() {
	s.btn = map[string]changeFunc{
		filterSetting:       s.nextSettingPageBtn,
		btnCancel:           s.cancelBtn,
		btnOk:               s.okBtn,
		btnDelete:           s.deleteBtn,
		changeStateBtn:      s.changeStateBtn,
		btnGetOldApartments: s.getOldApartmentsBtn,
		btnGetNewApartments: s.getNewApartmentsBtn,
	}
}

func (s *service) initHandlers() {
	s.b.Use(s.errorMiddleware)
	s.b.Handle("/start", s.startChatHandler)
	s.b.Handle(filterCommand, s.startCreatingFilterHandler)
	s.b.Handle("/get_filters", s.filtersListHandler)
	s.b.Handle("/help", s.helpHandler)
	s.b.Handle(tele.OnCallback, s.callbackHandler)
	s.b.Handle(tele.OnText, s.messageHandler)
	s.b.Handle(tele.OnLocation, s.locationHandler)
}

func (s *service) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.sendingMessage(s.ctx)
	go s.apartmentRuntime(s.service.Watcher())
	go s.b.Start()

	return nil
}

func (s *service) Stop() {
	s.cancel()
	s.b.Stop()
}

func (s *service) apartmentRuntime(apartmentCh <-chan server.Apartment) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case a, ok := <-apartmentCh:
			if !ok {
				return
			}
			s.sendApartment(a)
		}
	}
}

func (s *service) sendingMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-s.sendMessageCh:
			_, err := s.sendMessageToBot(m.UserID, m.What, m.Opts...)
			if err != nil {
				s.handleError(m.UserID, err)
			}

			interval := s.messageSendInterval
			if err != nil {
				interval = s.extractRetryTime(err.Error())
				slog.Info("sendApartmentAlbum", "interval", interval)
			}
			<-time.After(interval)
		}
	}
}

func (s *service) startChatHandler(c tele.Context) error {
	err := s.service.StartChat(s.ctx, userFromContext(c))
	if err != nil {
		return err
	}

	_, err = s.sendMessageToBot(c.Sender().ID, startChatMessage)
	if err != nil {
		return err
	}

	return s.filtersListHandler(c)
}

func (s *service) startCreatingFilterHandler(c tele.Context) error {
	userID := c.Sender().ID

	err := s.messages.CleanUserMessages(userID)
	if err != nil {
		return err
	}

	s.userAction.Delete(userID)

	_, err = s.sendMessageToBot(userID, creatingFilterMessage)
	if err != nil {
		return err
	}

	f := s.service.StartCreatingFilter(s.ctx, userFromContext(c))

	return s.sendSettingFilter(c, f)
}

func (s *service) filtersListHandler(c tele.Context) error {
	filters, err := s.service.Filters(s.ctx, userFromContext(c))
	if err != nil {
		slog.Error("get_filters", "err", err)
		return nil
	}

	err = s.messages.CleanUserMessages(c.Sender().ID)
	if err != nil {
		return err
	}
	s.userAction.Delete(c.Sender().ID)

	msg := filterListIsEmptyMessage
	if len(filters) > 0 {
		msg = "A list of your filters is available\n" + filtersStr(filters)
	}

	m, err := s.sendMessageToBot(c.Sender().ID, msg, s.filterMenu(filters))
	if err != nil {
		return err
	}
	s.messages.StoreMessage(c.Chat().ID, m, settingFilterMessage)

	return nil
}

func (s *service) helpHandler(c tele.Context) error {
	m, err := s.sendMessageToBot(c.Sender().ID, fmt.Sprintf(helpMessage, s.adminUsername))
	if err != nil {
		return err
	}

	s.messages.StoreMessage(c.Chat().ID, m, botMessage)
	return nil
}

func (s *service) callbackHandler(c tele.Context) error {
	userID := c.Sender().ID
	actionType := getType(c)
	actionValue := getValue(c)

	a, isExist := s.userAction.Load(userID)
	if isExist && a.(string) == actionType && len(actionValue) != 0 && actionValue[0] != nextPage { // nolint:errcheck
		return s.params[actionType].change(c)
	}

	if btn, isExist := s.btn[actionType]; isExist {
		err := btn(c)
		if err != nil {
			return err
		}
		return nil
	}

	if err := s.params[actionType].init(c); err != nil {
		return err
	}

	return nil
}

func (s *service) messageHandler(c tele.Context) error {
	userID := c.Sender().ID

	s.messages.StoreMessage(userID, c.Message(), userMassage)

	action, isExist := s.userAction.Load(userID)
	if !isExist {
		return s.chooseFilter(c)
	}

	return s.params[action.(string)].change(c) // nolint: errcheck
}

func (s *service) locationHandler(c tele.Context) error {
	userID := c.Sender().ID

	s.messages.StoreMessage(userID, c.Message(), userMassage)

	action, isExist := s.userAction.Load(userID)
	if isExist {
		return s.params[action.(string)].change(c) // nolint: errcheck
	}

	m, err := s.sendMessageToBot(c.Sender().ID, unknownCommandMessage)
	if err != nil {
		return err
	}
	s.messages.StoreMessage(userID, m, botMessage)
	return nil
}

func (s *service) chooseFilter(c tele.Context) error {
	userID := c.Sender().ID
	name := c.Text()
	f := &server.Filter{
		Name: &name,
		User: &server.User{
			ID: userID,
		},
	}

	filter, err := s.service.Filter(s.ctx, f)
	if err != nil {
		return err
	}

	err = s.messages.CleanUserMessages(userID)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}

func (s *service) sendMessageToBot(userID int64, what interface{}, opts ...interface{}) (*tele.Message, error) {
	m := Message{
		UserID: userID,
		What:   what,
		Opts:   opts,
		Answer: make(chan answer),
	}

	s.sendMessageCh <- m
	answer := <-m.Answer

	return answer.m, answer.err
}
