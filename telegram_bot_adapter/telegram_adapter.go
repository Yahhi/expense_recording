// Package telegram initializes the bot and applies interface [Bot] to actual telegram functionality
package telegram

import (
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"ingresos_gastos/bot_interface"
	"ingresos_gastos/config"
	"ingresos_gastos/storage_interface"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type BotAdapter struct {
	Bot     *telebot.Bot
	Storage storage_interface.ActualStorage
}

type TelegramEditable struct {
	Message   storage_interface.Message
	Recipient bot_interface.BotRecipient
}

func telebotUserToInterface(user *telebot.User) bot_interface.BotRecipient {
	return bot_interface.BotRecipient{UserID: user.ID, Name: user.Username}
}

func (message TelegramEditable) MessageSig() (messageID string, chatID int64) {
	return message.Message.ID, message.Recipient.UserID
}

func convertToOptions(items []bot_interface.Option, withFinishOption bool) *telebot.ReplyMarkup {
	heightOfArray := int(math.Ceil(float64(len(items)) / float64(3)))
	if withFinishOption {
		heightOfArray += 1
	}
	options := make([][]telebot.InlineButton, heightOfArray)
	for i := range options {
		options[i] = make([]telebot.InlineButton, 3)
	}
	for rowIndex, element := range items {
		options[rowIndex/3][rowIndex%3] = telebot.InlineButton{Unique: "inline_" + element.Id, Text: element.Text}
	}
	rowIndex := len(items)
	if rowIndex%3 == 1 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0]}
	} else if rowIndex%3 == 2 {
		options[rowIndex/3] = []telebot.InlineButton{options[rowIndex/3][0], options[rowIndex/3][1]}
	}
	options[heightOfArray-1] = []telebot.InlineButton{{Unique: "inline_" + bot_interface.CommandCancel, Text: "\xE2\x9B\x94Finish action"}}
	return &telebot.ReplyMarkup{InlineKeyboard: options}
}

func NewBotAdapter(cfg config.Config, storage storage_interface.ActualStorage) (BotAdapter, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TgBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second}, //todo to OS.ENV
	})

	commands := []telebot.Command{
		{Text: bot_interface.CommandStart, Description: "Hello"},
		{Text: bot_interface.CommandHelp, Description: "List of all commands"},
		{Text: bot_interface.CommandDefineTags, Description: "Define categories of expenses"},
		{Text: bot_interface.CommandDefineBudget, Description: "Set a budget for the current month"},
		{Text: bot_interface.CommandStatistics, Description: "View your statistics"},
		{Text: bot_interface.CommandFeedback, Description: "Describe your experience"},
		{Text: bot_interface.CommandCancel, Description: "Cancel current action"},
	}
	errSettingCommand := bot.SetCommands(commands)
	if errSettingCommand != nil {
		log.Print(fmt.Errorf("error setting commands for telegram bot: %v", errSettingCommand))
	}

	bot.Start()

	return BotAdapter{Bot: bot, Storage: storage}, err
}

func (adapter BotAdapter) Send(recipient bot_interface.BotRecipient, messages []bot_interface.Message) error {
	oldMessages, errorGettingMessages := adapter.Storage.GetMessages(recipient.UserID)
	if errorGettingMessages == nil {
		adapter.delete(recipient, oldMessages)
	} else {
		log.Print(fmt.Errorf("error getting messages from DB in Send: %v", errorGettingMessages))
	}

	for _, message := range messages {
		optionsLength := len(message.Options)
		var sentMessage *telebot.Message
		var err error
		if optionsLength > 0 {
			sentMessage, err = adapter.Bot.Send(recipient, message.Text, convertToOptions(message.Options, false))
			if err != nil {
				log.Print(fmt.Errorf("error sending message in Send: %v", err))
				return err
			}
		} else {
			sentMessage, err = adapter.Bot.Send(recipient, message.Text)
			if err != nil {
				log.Print(fmt.Errorf("error sending message in Send: %v", err))
				return err
			}
		}
		errSaving := adapter.Storage.SaveMessage(storage_interface.Message{ID: strconv.Itoa(sentMessage.ID), UserID: recipient.Recipient()})
		if errSaving != nil {
			log.Print(fmt.Errorf("error saving message in Send: %v", err))
		}
	}
	return nil
}

func (adapter BotAdapter) delete(recipient bot_interface.BotRecipient, messages []storage_interface.Message) {
	for _, message := range messages {
		err := adapter.Bot.Delete(TelegramEditable{Message: message, Recipient: recipient})
		if err != nil {
			log.Print(fmt.Errorf("error deleting messages in chat while Delete: %v", err))
		}
	}
	errInDB := adapter.Storage.ClearOutgoingMessagesForUser(recipient.UserID)
	if errInDB != nil {
		log.Print(fmt.Errorf("error removing messages in Database while Delete: %v", errInDB))
	}
}

func (adapter BotAdapter) ListenToCommand(command string, action func(recipient bot_interface.BotRecipient) ([]bot_interface.Message, error)) {
	adapter.Bot.Handle(command, func(m *telebot.Message) {
		recipient := bot_interface.BotRecipient{UserID: m.Sender.ID, Name: m.Sender.Username}
		messages, err := action(recipient)
		if err != nil {
			log.Print(fmt.Errorf("error building messages on command %s: %v", command, err))
			return
		}
		errSending := adapter.Send(recipient, messages)
		if errSending != nil {
			log.Print(fmt.Errorf("error sending messages on command %s: %v", command, errSending))
			return
		}
	})
}

// ListenToInput handles user's text input
func (adapter BotAdapter) ListenToInput(action func(recipient bot_interface.BotRecipient, text string) ([]bot_interface.Message, error)) {
	adapter.Bot.Handle(telebot.OnText, func(message *telebot.Message) {
		recipient := bot_interface.BotRecipient{UserID: message.Sender.ID, Name: message.Sender.Username}
		messages, err := action(recipient, message.Text)
		if err != nil {
			log.Print(fmt.Errorf("error getting messages for user's text input %s: %v", message.Text, err))
			return
		}
		errSending := adapter.Send(recipient, messages)
		if errSending != nil {
			log.Print(fmt.Errorf("error sending messages in reply to %s: %v", message.Text, errSending))
			return
		}
	})
}

// ListenToInlineActions handles inline buttons
func (adapter BotAdapter) ListenToInlineActions(action func(recipient bot_interface.BotRecipient, inlineAction string) ([]bot_interface.Message, error)) {
	adapter.Bot.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		inlineCommand := strings.TrimSpace(c.Data)
		recipient := bot_interface.BotRecipient{UserID: c.Sender.ID, Name: c.Sender.Username}
		if strings.HasPrefix(inlineCommand, "inline_") {
			tagName := strings.TrimPrefix(inlineCommand, "inline_")
			messages, err := action(recipient, tagName)
			if err != nil {
				log.Print(fmt.Errorf("error getting messages for inlineAction %s: %v", inlineCommand, err))
				return
			}
			errSending := adapter.Send(recipient, messages)
			if errSending != nil {
				log.Print(fmt.Errorf("error sending messages in reply to inlineAction: %v", inlineCommand, errSending))
				return
			}
		} else {
			//there is strange input. Log it!
			log.Print("Strange input " + inlineCommand + " from " + c.Sender.Username)
		}
	})
}
