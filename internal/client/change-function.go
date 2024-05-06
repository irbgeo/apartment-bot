package client

import (
	"context"
	"fmt"
	"time"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var defaultDistanceToLocation = 5000.0

func (s *service) ChangeFilterName(ctx context.Context, i *ChangeFilterNameInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.Name = i.NewName

	f := server.Filter{
		User: &server.User{
			ID: i.User.ID,
		},
		Name: i.NewName,
	}
	filter, err := s.srv.Filter(ctx, f)
	if err != nil && err.Error() != ErrFilterNotFound.Error() {
		return nil, err
	}
	if filter != nil {
		*i.ActiveFilter = *filter
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeTypeFilter(ctx context.Context, i *ChangeAdTypeFilterInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.AdType = i.NewAdType

	return i.ActiveFilter, nil
}

func (s *service) ChangeBuildingStatusFilter(ctx context.Context, i *ChangeBuildingStatusFilterInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.BuildingStatus = i.NewBuildingStatus

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterCity(ctx context.Context, i *ChangeFilterCityInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.City = i.NewCity
	i.ActiveFilter.District = make(map[string]struct{})

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterDistrict(ctx context.Context, i *ChangeFilterDistrictInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	if i.ChoseDistrict == nil {
		i.ActiveFilter.District = make(map[string]struct{})
	} else {
		_, ok := i.ActiveFilter.District[*i.ChoseDistrict]
		if ok {
			delete(i.ActiveFilter.District, *i.ChoseDistrict)
		} else {
			i.ActiveFilter.District[*i.ChoseDistrict] = struct{}{}
		}
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterPrice(ctx context.Context, i *ChangeFilterPriceInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	if i.IsMinChange {
		i.ActiveFilter.MinPrice = i.NewMinPrice
	} else {
		i.ActiveFilter.MaxPrice = i.NewMaxPrice
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterRooms(ctx context.Context, i *ChangeFilterRoomsInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	if i.IsMinChange {
		i.ActiveFilter.MinRooms = i.NewMinRooms
	} else {
		i.ActiveFilter.MaxRooms = i.NewMaxRooms
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterArea(ctx context.Context, i *ChangeFilterAreaInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	if i.IsMinChange {
		i.ActiveFilter.MinArea = i.NewMinArea
	} else {
		i.ActiveFilter.MaxArea = i.NewMaxArea
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterLocation(ctx context.Context, i *ChangeFilterLocationInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.Coordinates = nil

	if i.NewCoordinates != nil {
		i.ActiveFilter.Coordinates = &server.Coordinates{
			Lat: i.NewCoordinates.Lat,
			Lng: i.NewCoordinates.Lng,
		}
		i.ActiveFilter.MaxDistance = &defaultDistanceToLocation
	}

	fmt.Println(i.ActiveFilter.Coordinates)

	if i.NewCoordinates == nil {
		i.ActiveFilter.MaxDistance = nil
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeFilterMaxDistance(ctx context.Context, i *ChangeFilterMaxDistanceInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.MaxDistance = i.NewMaxDistance
	if i.NewMaxDistance == nil {
		i.ActiveFilter.Coordinates = nil
	}

	return i.ActiveFilter, nil
}

func (s *service) ChangeStateFilter(ctx context.Context, i *ChangeStateFilterInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	if i.ActiveFilter.PauseTimestamp == nil {
		date := time.Now().Unix()
		i.ActiveFilter.PauseTimestamp = &date

		return i.ActiveFilter, nil
	}

	i.ActiveFilter.PauseTimestamp = nil

	return i.ActiveFilter, nil
}

func (s *service) ChangeOwnerTypeFilter(ctx context.Context, i *ChangeOwnerTypeFilterInfo) (*server.Filter, error) {
	i.ActiveFilter.IsUpdate = true

	i.ActiveFilter.IsOwner = i.NewOwnerType

	return i.ActiveFilter, nil
}
