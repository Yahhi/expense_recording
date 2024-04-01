// Package bot_interface is an interface to make abstraction above actual telegram or another messenger interface.
// It makes possible to test actions of actual program making calls to known methods and getting
// [Message] answers
package bot_interface

import (
	"strconv"
)

type Bot interface {
	Send(recipient BotRecipient, messages []Message) error
	ListenToCommand(command string, action func(recipient BotRecipient) ([]Message, error))
	ListenToInput(action func(recipient BotRecipient, text string) ([]Message, error))
	ListenToInlineActions(action func(recipient BotRecipient, inlineAction string) ([]Message, error))
}

type BotRecipient struct {
	UserID int64
	Name   string
}

func (user BotRecipient) Recipient() string {
	return strconv.Itoa(int(user.UserID))
}

type Message struct {
	Text    string
	Id      string
	Options []Option
}

type Option struct {
	Id   string
	Text string
}

const (
	StateCreateTags   = "tag_create"
	StateModifyBudget = "tag_budget"
	StateSpending     = "tag_spending"
	StateFeedback     = "tag_spending"

	CommandCancel       = "cancel"
	CommandStart        = "start"
	CommandHelp         = "help"
	CommandDefineTags   = "define_tags"
	CommandDefineBudget = "define_budget"
	CommandStatistics   = "view_statistics"
	CommandFeedback     = "feedback"
)
