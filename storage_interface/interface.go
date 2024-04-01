// Package storage_interface defines the standard to which any database connection should fit to be appropriate to work
// with the bot_interface
// There are also structure definitions for data used in bot_interface
package storage_interface

import (
	"time"
)

type ActualStorage interface {
	CreateUser(userID int64, name string) error
	UserExists(userID int64) (bool, error)
	SetState(userID int64, state string) error
	GetUserState(userID int64) (string, error)

	CreateTarget(tag string, amount float32, periodStart time.Time, periodEnd time.Time, userID int64) error
	GetTargets(periodStart time.Time, periodEnd time.Time, userID int64) ([]Target, error)

	CreateMoneyEvent(amount float32, currency, comment, tag string, userID int64) error
	GetMoneyEventsByDateInterval(startDate, endDate time.Time, userID int64) ([]MoneyEvent, error)

	AddTagForUser(tag string, userID int64) error
	RemoveTagForUser(tag string, userID int64) error
	GetUserTags(userID int64) ([]string, error)

	SaveFeedback(userID int64, message string) error

	SaveMessage(message Message) error
	GetMessages(userId int64) ([]Message, error)
	ClearOutgoingMessagesForUser(userID int64) error

	SaveUsageLog(userId int64, replyType string) error
}

// User is a telegram user, who once spoke with the bot_interface
type User struct {
	ID      int
	Name    string
	State   string
	Created time.Time
}

// Target describes how much money the [User] wants to spend in the period for a special spending tag.
// Usually it is about current month
type Target struct {
	ID          int
	Tag         string
	Amount      float32
	PeriodStart time.Time
	PeriodEnd   time.Time
	UserID      int
}

// MoneyEvent is a spending event. It happens when [User] spends some money in a cafe or buys something
// and tells this fact to the bot_interface
type MoneyEvent struct {
	ID       int
	Amount   float32
	Currency string
	Comment  string
	Tag      string
	Created  time.Time
	UserID   int
}

// Message is
type Message struct {
	ID     string
	UserID string
}
