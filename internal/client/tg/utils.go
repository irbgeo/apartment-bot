package tg

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/irbgeo/apartment-bot/internal/server"
)

func (s *service) cleanUserMessages(userID int64) error {
	return s.messages.CleanUserMessages(userID)
}

func filtersStr(filters []server.Filter) string {
	var result strings.Builder
	for _, f := range filters {
		result.WriteString("#" + *f.Name + "\n")
	}
	return result.String()
}

func userFromContext(c tele.Context) *server.User {
	return &server.User{
		ID: c.Sender().ID,
	}
}

func apartmentString(a server.Apartment, filters []string) string {
	var hashtags strings.Builder
	for _, name := range filters {
		hashtags.WriteString("#" + name + "\n")
	}

	var comment string
	if len(a.Comment) > 0 {
		comment = a.Comment[:min(50, len(a.Comment))] + "..."
		comment = "\nComment: " + comment
	}

	location := ""
	if a.Coordinates != nil {
		location = locationString(a.Coordinates.Lat, a.Coordinates.Lng)
	}

	year, month, day := a.OrderDate.Date()
	return fmt.Sprintf(
		apartmentStrTemplate,
		hashtags.String(),
		a.URL,
		typeMap[a.AdType],
		ownerTypeMap[a.IsOwner],
		a.Price,
		a.Phone,
		a.Rooms,
		a.Bedrooms,
		a.Floor,
		a.Area,
		a.District,
		a.City,
		location,
		comment,
		day, month, year,
	)
}

func actionData(args ...string) string {
	return strings.Join(args, dataSep)
}

func locationString(lat, lng float64) string {
	return fmt.Sprintf("%s%s,%s", locationURL, strconv.FormatFloat(lat, 'f', -1, 64), strconv.FormatFloat(lng, 'f', -1, 64))
}

func getType(c tele.Context) string {
	if callback := c.Callback(); callback != nil {
		return strings.Split(callback.Data, dataSep)[0]
	}
	return ""
}

func getValue(c tele.Context) []string {
	if callback := c.Callback(); callback != nil {
		parts := strings.Split(callback.Data, dataSep)
		if len(parts) >= 2 {
			return parts[1:]
		}
	}
	return nil
}

var (
	horizontalN = 4
	verticalN   = 3
	groupeN     = horizontalN * verticalN
)

func group(str []string) [][]string {
	groups := make([][]string, 0, (len(str)+groupeN-1)/groupeN)

	for i := 0; i < len(str); i += groupeN {
		end := min(i+groupeN, len(str))
		groups = append(groups, str[i:end])
	}

	return groups
}

func replaceURLSingleSlash(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "//", "/"), "/", "//")
}

func getImageSizeMB(url string) (float64, error) {
	resp, err := http.Head(url) // nolint: gosec
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size == -1 {
		return 0, fmt.Errorf("server did not provide Content-Length")
	}

	return float64(size) / (1024 * 1024), nil
}

func nextPageIdx(currentPage, numberOfPages int) int {
	return (currentPage + 1) % numberOfPages
}
