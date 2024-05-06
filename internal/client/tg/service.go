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
	ctx    context.Context
	cancel context.CancelFunc
	b      *tele.Bot

	message             messageSvc
	service             apartmentSvc
	adminUsername       string
	maxPhotoCount       int
	messageSendInterval time.Duration

	userAction sync.Map
	messages   messageStack

	params      map[string]param
	btn         map[string]changeFunc
	settingBtns [][][]func(f *server.Filter) tele.Btn
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

//go:generate mockery --name messageSvc --structname MessageSvc
type messageSvc interface {
	StartWatcher(ctx context.Context) (<-chan Message, <-chan error, error)
}

//go:generate mockery --name messageStack --structname MessageStack
type messageStack interface {
	SetBot(b *tele.Bot)
	Store(userID int64, m *tele.Message, t MessageType)
	GetOrCleanTill(userID int64, goalType, tillType MessageType) (*tele.Message, bool, error)
	Clean(userID int64) error
	CleanTill(userID int64, t MessageType) error
}

func NewService(
	token string,
	aSvc apartmentSvc,
	mSvc messageSvc,
	mStack messageStack,
	disabledParameters []string,
	adminUsername string,
	maxPhotoCount int,
	messageSendInterval time.Duration,
) (*service, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: botTimeout},
	})
	if err != nil {
		return nil, err
	}

	mStack.SetBot(b)

	t := &service{
		b:        b,
		messages: mStack,

		message:             mSvc,
		service:             aSvc,
		adminUsername:       adminUsername,
		maxPhotoCount:       maxPhotoCount,
		messageSendInterval: messageSendInterval,
	}

	t.params = map[string]param{
		changeName: {
			init:     t.changeNameInit,
			change:   t.changeName,
			toString: t.nameParamToString,
		},

		changeAdType: {
			init:     t.changeTypeInit,
			change:   t.changeAdType,
			toString: t.adTypeParamToString,
		},

		changeBuildingStatus: {
			init:     t.changeBuildingStatusInit,
			change:   t.changeBuildingStatus,
			toString: t.buildingStatusParamToString,
		},

		changeCity: {
			init:     t.changeCityInit,
			change:   t.changeCity,
			toString: t.cityParamToString,
		},

		changeDistrict: {
			init:     t.changeDistrictInit,
			change:   t.changeDistrict,
			toString: t.districtParamToString,
		},

		changeMinPrice: {
			init:     t.changePriceInit(true),
			change:   t.changePrice(true),
			toString: t.priceParamToString,
		},

		changeMaxPrice: {
			init:   t.changePriceInit(false),
			change: t.changePrice(false),
		},

		changeMinRooms: {
			init:     t.changeRoomsInit(true),
			change:   t.changeRooms(true),
			toString: t.roomsParamToString,
		},

		changeMaxRooms: {
			init:   t.changeRoomsInit(false),
			change: t.changeRooms(false),
		},

		changeMinArea: {
			init:     t.changeAreaInit(true),
			change:   t.changeArea(true),
			toString: t.areaParamToString,
		},

		changeMaxArea: {
			init:   t.changeAreaInit(false),
			change: t.changeArea(false),
		},

		changeLocation: {
			init:     t.changeLocationInit,
			change:   t.changeLocation,
			toString: t.locationParamToString,
		},

		changeMaxDistance: {
			init:     t.changeMaxDistanceInit,
			change:   t.changeMaxDistance,
			toString: t.maxDistanceParamToString,
		},

		changeOwnerType: {
			init:     t.changeOwnerTypeInit,
			change:   t.changeOwnerType,
			toString: t.ownerTypeParamToString,
		},
	}

	t.settingBtns = [][][]func(f *server.Filter) tele.Btn{
		{
			{changeNameBtn},
			{changeCityBtn, t.changeDistrictBtn},
			{t.changePriceBtn(true), t.changePriceBtn(false)},
			{t.changeAreaBtn(true), t.changeAreaBtn(false)},
			{changeLocationBtn, changeMaxDistanceBtn},
		},
		{
			{t.changeRoomsBtn(true), t.changeRoomsBtn(false)},
			{t.changeAdTypeBtn, t.changeOwnerTypeBtn},
			{t.changeBuildingStatusBtn},
		},
	}

	for _, param := range disabledParameters {
		delete(t.params, param)
	}

	t.btn = map[string]changeFunc{
		filterSetting:       t.nextSettingPageBtn,
		btnCancel:           t.cancelBtn,
		btnOk:               t.okBtn,
		btnDelete:           t.deleteBtn,
		changeStateBtn:      t.changeStateBtn,
		btnGetOldApartments: t.getOldApartmentsBtn,
		btnGetNewApartments: t.getNewApartmentsBtn,
	}

	b.Use(t.errorMiddleware)
	b.Handle("/start", t.startChatHandler)
	b.Handle(filterCommand, t.startCreatingFilterHandler)
	b.Handle("/get_filters", t.filtersListHandler)
	b.Handle("/help", t.helpHandler)

	b.Handle(tele.OnCallback, t.callbackHandler)
	b.Handle(tele.OnText, t.messageHandler)
	b.Handle(tele.OnLocation, t.locationHandler)

	return t, nil
}

func (s *service) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	messageCh, errCh, err := s.message.StartWatcher(s.ctx)
	if err != nil {
		return err
	}
	go s.messageRuntime(
		messageCh,
		errCh,
	)

	apartmentCh := s.service.Watcher()
	go s.apartmentRuntime(
		apartmentCh,
	)
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

func (s *service) messageRuntime(messageCh <-chan Message, errCh <-chan error) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-errCh:
			slog.Error("message_runtime", "err", errCh)
		case m, ok := <-messageCh:
			if !ok {
				return
			}
			s.sendPinMessage(m)
		}
	}
}

func (s *service) startChatHandler(c tele.Context) error {
	err := s.service.StartChat(s.ctx, userFromContext(c))
	if err != nil {
		return err
	}

	err = c.Send(startChatMessage)
	if err != nil {
		return err
	}

	return s.filtersListHandler(c)
}

func (s *service) startCreatingFilterHandler(c tele.Context) error {
	userID := c.Sender().ID

	err := s.messages.Clean(userID)
	if err != nil {
		return err
	}

	s.userAction.Delete(userID)

	err = c.Send(creatingFilterInfoMessage)
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

	err = s.messages.Clean(c.Sender().ID)
	if err != nil {
		return err
	}
	s.userAction.Delete(c.Sender().ID)

	msg := filterListIsEmptyMessage
	if len(filters) > 0 {
		msg = "A list of your filters is available\n" + filtersStr(filters)
	}

	m, err := s.b.Send(c.Sender(), msg, s.filterMenu(filters))
	if err != nil {
		return err
	}
	s.messages.Store(c.Chat().ID, m, settingFilterMessage)

	return nil
}

func (s *service) helpHandler(c tele.Context) error {
	m, err := s.b.Send(c.Sender(), fmt.Sprintf(helpMessage, s.adminUsername))
	if err != nil {
		return err
	}

	s.messages.Store(c.Chat().ID, m, botMessage)
	return nil
}

func (s *service) callbackHandler(c tele.Context) error {
	userID := c.Sender().ID
	actionType := getType(c)
	actionValue := getValue(c)

	a, isExist := s.userAction.Load(userID)
	if isExist && a.(string) == actionType && len(actionValue) != 0 && actionValue[0] != nextPage {
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

	s.messages.Store(userID, c.Message(), userMassage)

	action, isExist := s.userAction.Load(userID)
	if !isExist {
		return s.chooseFilter(c)
	}

	return s.params[action.(string)].change(c)
}

func (s *service) locationHandler(c tele.Context) error {
	userID := c.Sender().ID

	s.messages.Store(userID, c.Message(), userMassage)

	action, isExist := s.userAction.Load(userID)
	if isExist {
		return s.params[action.(string)].change(c)
	}

	m, err := s.b.Send(c.Sender(), unknownCommandMessage)
	if err != nil {
		return err
	}
	s.messages.Store(userID, m, botMessage)
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

	err = s.messages.Clean(userID)
	if err != nil {
		return err
	}

	s.userAction.Delete(c.Sender().ID)

	return s.sendSettingFilter(c, filter)
}
