package ssge

import "time"

type StartOpts struct {
	RefreshTokenInterval time.Duration
}

type apartmentList struct {
	Data []data `json:"realStateItemModel"`
}

type data struct {
	ApplicationID int64 `json:"applicationId"`
}

type address struct {
	CityTitle        string `json:"cityTitle"`
	SubdistrictTitle string `json:"subdistrictTitle"`
}

type price struct {
	PriceUSD float64 `json:"priceUsd"`
}

type appImage struct {
	FileName      string `json:"fileName"`
	IsMain        bool   `json:"isMain"`
	Is360         bool   `json:"is360"`
	OrderNo       int64  `json:"orderNo"`
	ImageType     int64  `json:"imageType"`
	FileNameThumb string `json:"fileNameThumb"`
}

type applicationPhone struct {
	PhoneNumber string `json:"phoneNumber"`
}

type description struct {
	En string `json:"en"`
}

type apartment struct {
	ApplicationID         int64              `json:"applicationId"`
	IsInactiveApplication bool               `json:"isInactiveApplication"`
	RealEstateDealTypeID  int64              `json:"realEstateDealTypeId"`
	RealEstateStatusID    int64              `json:"realEstateStatusId"`
	Address               address            `json:"address"`
	Price                 price              `json:"price"`
	AppImages             []appImage         `json:"appImages"`
	ApplicationPhones     []applicationPhone `json:"applicationPhones"`
	Description           description        `json:"description"`
	BuildingStatus        string             `json:"status"`
	OrderDate             time.Time          `json:"orderDate"`
	LocationLatitude      float64            `json:"locationLatitude"`
	LocationLongitude     float64            `json:"locationLongitude"`
	Bedrooms              int64              `json:"bedrooms"`
	Floor                 string             `json:"floor"`
	Rooms                 string             `json:"rooms"`
	TotalArea             string             `json:"totalArea"`
	PriceLevel            string             `json:"priceLevel"`
	UserEntityType        string             `json:"userEntityType"`
	State                 string             `json:"state"`
}

type requestApartmentListBody struct {
	AdvancedSearch advancedSearch `json:"advancedSearch"`
	RealEstateType int64          `json:"realEstateType"`
	CurrencyID     int64          `json:"currencyId"`
	Order          int64          `json:"order"`
	Page           int64          `json:"page"`
	PageSize       int64          `json:"pageSize"`
}

type advancedSearch struct {
	WithImageOnly bool `json:"withImageOnly"`
}
