package storage_interface

import "time"

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
}

type User struct {
	ID      int
	Name    string
	State   string
	Created time.Time
}

type Target struct {
	ID          int
	Tag         string
	Amount      float32
	PeriodStart time.Time
	PeriodEnd   time.Time
	UserID      int
}

type MoneyEvent struct {
	ID       int
	Amount   float32
	Currency string
	Comment  string
	Tag      string
	Created  time.Time
	UserID   int
}
