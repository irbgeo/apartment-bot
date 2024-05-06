package client

import "github.com/irbgeo/apartment-bot/internal/server"

type ChangeFilterNameInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	NewName      *string
}

func (s *ChangeFilterNameInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterNameInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeAdTypeFilterInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	NewAdType    *int64
}

func (s *ChangeAdTypeFilterInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeAdTypeFilterInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeBuildingStatusFilterInfo struct {
	User              *server.User
	ActiveFilter      *server.Filter
	NewBuildingStatus *int64
}

func (s *ChangeBuildingStatusFilterInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeBuildingStatusFilterInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterCityInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	NewCity      *string
}

func (s *ChangeFilterCityInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterCityInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterDistrictInfo struct {
	User          *server.User
	ActiveFilter  *server.Filter
	ChoseDistrict *string
}

func (s *ChangeFilterDistrictInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterDistrictInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterPriceInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	IsMinChange  bool
	NewMinPrice  *float64
	NewMaxPrice  *float64
}

func (s *ChangeFilterPriceInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterPriceInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterRoomsInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	IsMinChange  bool
	NewMaxRooms  *float64
	NewMinRooms  *float64
}

func (s *ChangeFilterRoomsInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterRoomsInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterAreaInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	IsMinChange  bool
	NewMinArea   *float64
	NewMaxArea   *float64
}

func (s *ChangeFilterAreaInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterAreaInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterLocationInfo struct {
	User           *server.User
	ActiveFilter   *server.Filter
	NewCoordinates *Coordinates
}

type Coordinates struct {
	Lat float64
	Lng float64
}

func (s *ChangeFilterLocationInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterLocationInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeFilterMaxDistanceInfo struct {
	User           *server.User
	ActiveFilter   *server.Filter
	NewMaxDistance *float64
}

func (s *ChangeFilterMaxDistanceInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeFilterMaxDistanceInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeStateFilterInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
}

func (s *ChangeStateFilterInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeStateFilterInfo) GetUserID() int64 {
	return s.User.ID
}

type ChangeOwnerTypeFilterInfo struct {
	User         *server.User
	ActiveFilter *server.Filter
	NewOwnerType *bool
}

func (s *ChangeOwnerTypeFilterInfo) SetActiveFilter(f *server.Filter) {
	s.ActiveFilter = f
}

func (s *ChangeOwnerTypeFilterInfo) GetUserID() int64 {
	return s.User.ID
}
