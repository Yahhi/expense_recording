package speaking

import (
	"ingresos_gastos/bot_interface"
	"strings"
	"time"
)

var defaultTags = []bot_interface.Option{
	{Id: "inline_Food", Text: "\xF0\x9F\x8D\x9CFood"},
	{Id: "inline_Cafe", Text: "\xE2\x98\x95Cafe"},
	{Id: "inline_Bar", Text: "\xF0\x9F\x8D\xB9Bar"},
	{Id: "inline_Auto", Text: "\xF0\x9F\x9A\x99Auto"},
	{Id: "inline_Medicine", Text: "\xF0\x9F\x9A\x91Medicine"},
	{Id: "inline_Credits", Text: "Credits"},
	{Id: "inline_Travel", Text: "\xF0\x9F\x9A\xA2Travel"},
	{Id: "inline_Garden", Text: "\xF0\x9F\x8C\xBCGarden"},
	{Id: "inline_Culture", Text: "\xF0\x9F\x8E\xADCulture"},
	{Id: "inline_Home", Text: "\xF0\x9F\x8F\xA0Home"},
	{Id: "inline_Pet", Text: "\xF0\x9F\x90\xA9Pet"},
	{Id: "inline_Clothes", Text: "\xF0\x9F\x91\x97Clothes"},
	{Id: "inline_Investment", Text: "\xF0\x9F\x92\x8EInvestment"}}

// gets interval for current month to compare with database dates
func monthInterval() (time.Time, time.Time) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	a := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	b := a.AddDate(0, 1, 0)
	return a, b
}

func trimStringFromFirstSpace(s string) string {
	if idx := strings.Index(s, " "); idx != -1 {
		return s[idx+1:]
	}
	return s
}
