package speaking

import (
	"math"
	"strings"
	"time"

	"gopkg.in/tucnak/telebot.v2"
)

var defaultTags = []string{"\xF0\x9F\x8D\x9CFood", "\xE2\x98\x95Cafe", "\xF0\x9F\x8D\xB9Bar", "\xF0\x9F\x9A\x99Auto", "\xF0\x9F\x9A\x91Medicine", "Credits", "\xF0\x9F\x9A\xA2Travel", "\xF0\x9F\x8C\xBCGarden", "\xF0\x9F\x8E\xADCulture", "\xF0\x9F\x8F\xA0Home", "\xF0\x9F\x90\xA9Pet", "\xF0\x9F\x91\x97Clothes", "\xF0\x9F\x92\x8EInvestment"}

func (env Env) convertMapToOptions(items map[string]string) *telebot.ReplyMarkup {
	heightOfArray := int(math.Ceil(float64(len(items)) / float64(3)))
	options := make([][]telebot.InlineButton, heightOfArray)
	for i := range options {
		options[i] = make([]telebot.InlineButton, 3)
	}
	rowIndex := 0
	for key, value := range items {
		options[rowIndex/3][rowIndex%3] = telebot.InlineButton{Unique: "inline_" + key, Text: value}
		rowIndex++
	}
	if rowIndex%3 == 1 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0]}
	} else if rowIndex%3 == 2 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0], options[rowIndex/3][1]}
	}
	return &telebot.ReplyMarkup{InlineKeyboard: options}
}

func (env Env) convertStringsToOptions(items []string, withFinishOption bool) *telebot.ReplyMarkup {
	heightOfArray := int(math.Ceil(float64(len(items)) / float64(3)))
	if withFinishOption {
		heightOfArray += 1
	}
	options := make([][]telebot.InlineButton, heightOfArray)
	for i := range options {
		options[i] = make([]telebot.InlineButton, 3)
	}
	for rowIndex, element := range items {
		options[rowIndex/3][rowIndex%3] = telebot.InlineButton{Unique: "inline_" + element, Text: element}
	}
	rowIndex := len(items)
	if rowIndex%3 == 1 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0]}
	} else if rowIndex%3 == 2 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0], options[rowIndex/3][1]}
	}
	options[heightOfArray-1] = []telebot.InlineButton{{Unique: "inline_" + CommandCancel, Text: "\xE2\x9B\x94Finish action"}}
	return &telebot.ReplyMarkup{InlineKeyboard: options}
}

// gets interval for current month to compare with database dates
func (env Env) monthInterval() (time.Time, time.Time) {
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
