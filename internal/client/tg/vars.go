package tg

import (
	"fmt"

	"github.com/irbgeo/apartment-bot/internal/server"
)

const (
	dataSep       = ":"
	filterCommand = "/create_filter"
	anyValue      = "any"
	unknownValue  = "unknown"
)

var (
	startChatMessage             = "Hello, I'm apartment bot!\nI will help you find an apartment in Georgia\n\n"
	notActiveFilterMessageLayout = "You don't have active filter.\nPlease, start creating it first: %s"
	unknownCommandMessage        = "What do you mean?"
	filterListIsEmptyMessage     = fmt.Sprintf(`You don't have any filters. You will not receive any apartments.
	If you want to start searching for apartments, create a filter: %s`, filterCommand)
	helpMessage = `Instructions: https://telegra.ph/Apartments-in-Georgia-bot-04-07
If you have any questions, contact us at @%s.`

	creatingFilterInfoMessage = `Let's start creating a filter for apartment hunting! Specify the parameters you need.

	To adjust apartment search parameters, click on ⚙️. Then press ✅ to save the filter and start receiving relevant listings. If you need to modify the criteria or delete the filter, select it from the menu below the message input field.
	
	You always can get help by using the command /help`

	locationURL = `📍 https://www.google.com/maps/search/?api=1&query=`
)

var (
	typeMap = map[int64]string{
		server.RentAdType: "Rent",
		server.SaleAdType: "Sale",
	}

	ownerTypeMap = map[bool]string{
		true:  "Owner",
		false: "Agency",
	}

	maxImageSizeMB = 2.0

	apartmentStrTemplate = `
%s
🌐 %s
Type: %s
From: %s

Price: %.1f$
☎️ +995%s

Rooms: %.0f
Bedrooms: %d
Floor: %d
Area: %.1f m2

District: %s
City: %s
%s

%s 

Date: %d %s %d
`
)
