package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/tucnak/telebot.v2"
	"ingresos_gastos/config"
	"ingresos_gastos/db"
	"ingresos_gastos/speaking"
)

// t@Gastos_Ingresos_bot
func main() {
	cfg := config.GetConfigFromEnv()
	fmt.Println(cfg)
	// Initialize the database
	storage := db.NewPostgresAdapter(cfg)

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TgBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second}, //todo to OS.ENV
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	env := speaking.Env{Storage: storage, Bot: bot}
	bot.Handle("/"+speaking.CommandStart, func(m *telebot.Message) {
		env.ProvideGreeting(m.Sender)
	})

	bot.Handle("/"+speaking.CommandHelp, func(m *telebot.Message) {
		env.ProvideHelp(m.Sender)
	})

	bot.Handle("/"+speaking.CommandDefineTags, func(m *telebot.Message) {
		env.GiveInstructionsOnTags(m.Sender)
	})

	bot.Handle("/"+speaking.CommandDefineBudget, func(m *telebot.Message) {
		env.GiveInstructionOnBudgeting(m.Sender)
	})

	bot.Handle("/"+speaking.CommandStatistics, func(m *telebot.Message) {
		env.GiveCurrentStatistics(m.Sender)
	})

	bot.Handle("/"+speaking.CommandCancel, func(m *telebot.Message) {
		env.CancelLastState(m.Sender)
	})

	//handle inline buttons
	bot.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		inlineCommand := strings.TrimSpace(c.Data)
		if strings.HasPrefix(inlineCommand, "inline_") {
			tagName := strings.TrimPrefix(inlineCommand, "inline_")
			env.DetectAppropriateActionForButton(c.Sender, tagName)
		} else {
			//there is strange input. Log it!
			log.Print("Strange input " + inlineCommand + " from " + c.Sender.Username)
		}
	})

	//handle text input
	bot.Handle(telebot.OnText, func(message *telebot.Message) {
		env.DetectAppropriateActionForInput(message.Sender, message.Text)
	})

	commands := []telebot.Command{
		{Text: speaking.CommandStart, Description: "Start the bot and get a description"},
		{Text: speaking.CommandHelp, Description: "Get a list of all available commands"},
		{Text: speaking.CommandDefineTags, Description: "Define categories of expenses"},
		{Text: speaking.CommandDefineBudget, Description: "Set a budget for each category for the current month"},
		{Text: speaking.CommandStatistics, Description: "View your current month's statistics"},
		{Text: speaking.CommandCancel, Description: "Cancel current action"},
	}
	errSettingCommand := bot.SetCommands(commands)
	if errSettingCommand != nil {
		log.Print(fmt.Errorf("error setting commands for bot: %v", errSettingCommand))
	}

	bot.Start()
}
