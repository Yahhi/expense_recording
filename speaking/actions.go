package speaking

import (
	"fmt"
	"ingresos_gastos/bot_interface"
	"ingresos_gastos/storage_interface"
	"log"
	"strings"
)

// MessagingPlatform contains full functionality to speak with users.
// There is an interface to store data [Storage] and an interface to
// provide changes in messages available to user in chat
type MessagingPlatform struct {
	Storage storage_interface.ActualStorage
	Bot     bot_interface.Bot
}

func provideMainOptions() bot_interface.Message {
	text := "To add new expense just type it here. Other commands:"
	options := []bot_interface.Option{
		{Id: bot_interface.CommandHelp, Text: "\xE2\x9D\x93help"},
		{Id: bot_interface.CommandDefineTags, Text: "\xE2\x9C\x8Ftags"},
		{Id: bot_interface.CommandDefineBudget, Text: "\xF0\x9F\x92\xB0budget"},
		{Id: bot_interface.CommandStatistics, Text: "\xF0\x9F\x93\x8Astatistics"},
	}
	return bot_interface.Message{
		Text:    text,
		Options: options,
	}
}

func (env MessagingPlatform) ProvideGreeting(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	reply := "👋 ¡Hola! Welcome to the Buenos Aires Expense Tracker Bot. I'm here to help you keep track of your daily expenses with ease. Whether you're trying to stay on budget or just want to see where your money goes, I've got you covered. Let's make managing your expenses a breeze! 💰📊"

	exists, errGettingUser := env.Storage.UserExists(user.UserID)
	if errGettingUser != nil {
		log.Print(fmt.Errorf("error checking user existance in ProvideGreeting: %v", errGettingUser))
	}
	if !exists {
		errInDb := env.Storage.CreateUser(user.UserID, user.Name)
		if errInDb != nil {
			log.Print(fmt.Errorf("error creating user in ProvideGreeting: %v", errInDb))
			return []bot_interface.Message{{Text: "Problem creating your profile in our system. Please try again later"}}, errInDb
		}
	}
	return []bot_interface.Message{{Text: reply}, provideMainOptions()}, nil
}

func (env MessagingPlatform) ProvideHelp(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	reply := fmt.Sprintf(`%s - Start the bot_interface and get a description
%s - Get a list of all available commands
%s - Define categories of expenses
%s - Set a budget for each category for the current month
<number> <tag> <comment> - save a new expense (only number is required, other fields are optional)
%s - View your current month statistics
%s - Give feedback to developers about this product
%s - finish ongoing operation (like defining month budget or creating tags)`,
		bot_interface.CommandStart, bot_interface.CommandHelp, bot_interface.CommandDefineTags,
		bot_interface.CommandDefineBudget, bot_interface.CommandStatistics, bot_interface.CommandFeedback,
		bot_interface.CommandCancel)
	return []bot_interface.Message{{Text: reply}, provideMainOptions()}, nil
}

func (env MessagingPlatform) ProvideFeedbackInstruction(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	reply := "📢 Your Feedback Matters!\n\nWe're always looking to improve your experience with the Buenos Aires Expense Tracker. If you have a moment, we'd love to hear your thoughts on how we can make this bot even better. Whether it's a new feature suggestion, a bug report, or just general feedback, we're all ears! Just send your feedback now!"
	err := env.Storage.SetState(user.UserID, bot_interface.StateFeedback)
	return []bot_interface.Message{{Text: reply}}, err
}

func (env MessagingPlatform) SaveFeedback(user bot_interface.BotRecipient, feedback string) ([]bot_interface.Message, error) {
	errSavingFeedback := env.Storage.SaveFeedback(user.UserID, feedback)
	if errSavingFeedback != nil {
		log.Print(fmt.Errorf("error saving feedback in SaveFeedback: %v", errSavingFeedback))
		return []bot_interface.Message{{Text: "Problem saving your data. Please try again later"}}, errSavingFeedback
	}
	err := env.Storage.SetState(user.UserID, "")
	if err != nil {
		log.Print(fmt.Errorf("error saving state in SaveFeedback: %v", err))
		return []bot_interface.Message{{Text: "Problem saving your data. Please try again later"}}, err
	}
	return []bot_interface.Message{{Text: "Thank you for helping us grow and serve you better!"}, provideMainOptions()}, nil
}

func (env MessagingPlatform) GiveInstructionsOnTags(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	acceptedTags, err := env.Storage.GetUserTags(user.UserID)
	if err != nil {
		log.Print(fmt.Errorf("error getting user tags in GiveInstructionsOnTags: %v", err))
		return []bot_interface.Message{{Text: "Problem creating your profile in our system. Please try again later"}}, err
	}
	var replyOptions []bot_interface.Option
	var textReply string
	if len(acceptedTags) == 0 {
		replyOptions = defaultTags
		textReply = "You didn't select tags yet. Please select from the list to add or input your own by keyboard"
	} else {
		for _, savedTag := range acceptedTags {
			replyOptions = append(replyOptions, bot_interface.Option{Id: "inline_" + savedTag, Text: savedTag})
		}
		textReply = "Your current tags are below.\nSelect a tag to delete or input new tags by keyboard"
	}
	errSavingState := env.Storage.SetState(user.UserID, bot_interface.StateCreateTags)
	if errSavingState != nil {
		log.Print(fmt.Errorf("error saving user state in GiveInstructionsOnTags: %v", errSavingState))
		return []bot_interface.Message{{Text: "Problem creating your profile in our system. Please try again later"}}, errSavingState
	}
	return []bot_interface.Message{{Text: textReply, Options: replyOptions}}, nil
}

func (env MessagingPlatform) CancelLastState(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	errSavingState := env.Storage.SetState(user.UserID, "")
	if errSavingState != nil {
		log.Print(fmt.Errorf("error saving user state in CancelLastState: %v", errSavingState))
		return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, errSavingState
	}
	return []bot_interface.Message{provideMainOptions()}, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (env MessagingPlatform) UpdateTag(user bot_interface.BotRecipient, tag string) ([]bot_interface.Message, error) {
	tags, err := env.Storage.GetUserTags(user.UserID)
	if err != nil {
		return nil, err
	}
	var reply string
	if stringInSlice(tag, tags) {
		err := env.Storage.RemoveTagForUser(tag, user.UserID)
		if err != nil {
			log.Print(fmt.Errorf("error removing tag in updateTag: %v", err))
			return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, err
		}
		reply = "Tag '" + tag + "' deleted"
	} else {
		err := env.Storage.AddTagForUser(tag, user.UserID)
		if err != nil {
			log.Print(fmt.Errorf("error adding tag in updateTag: %v", err))
			return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, err
		}
		reply = "Tag '" + tag + "' added"
	}
	return []bot_interface.Message{{Text: reply}, provideMainOptions()}, nil
}

func (env MessagingPlatform) GiveInstructionOnBudgeting(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	firstOfMonth, nextMonth := monthInterval()
	existingData, err := env.Storage.GetTargets(firstOfMonth, nextMonth, user.UserID)
	replyOptions := make(map[string]string)
	textReply := "To update budget for this month select a tag and then enter target amount"
	if err == nil {
		for _, existingData := range existingData {
			replyOptions[existingData.Tag] = fmt.Sprintf("%s - %.2f", existingData.Tag, existingData.Amount)
		}
	} else {
		log.Print(fmt.Errorf("error getting targets in GiveInstructionOnBudgeting: %v", err))
	}
	otherTags, errGettingTags := env.Storage.GetUserTags(user.UserID)
	if errGettingTags == nil {
		for _, tag := range otherTags {
			if _, ok := replyOptions[tag]; !ok {
				replyOptions[tag] = tag
			}
		}
	} else {
		log.Print(fmt.Errorf("error getting tags in GiveInstructionOnBudgeting: %v", errGettingTags))
	}

	var options []bot_interface.Option
	for key, value := range replyOptions {
		options = append(options, bot_interface.Option{Id: key, Text: value})
	}

	errSettingState := env.Storage.SetState(user.UserID, bot_interface.StateModifyBudget)
	if errSettingState != nil {
		log.Print(fmt.Errorf("error setting user state in GiveInstructionOnBudgeting: %v", errSettingState))
		return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, errSettingState
	}
	return []bot_interface.Message{{Text: textReply, Options: options}}, nil
}

func (env MessagingPlatform) ConfirmSelectingBudgetTag(user bot_interface.BotRecipient, tag string) ([]bot_interface.Message, error) {
	err := env.Storage.SetState(user.UserID, fmt.Sprintf("%s %s", bot_interface.StateModifyBudget, tag))
	if err != nil {
		log.Print(fmt.Errorf("error setting state in ConfirmSelectingBudgetTag: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, err
	}
	return []bot_interface.Message{{Text: "Enter updated amount of money you want to spend on '" + tag + "' in this month (or 0 if you don't want to spend money for this)"}}, nil
}

func (env MessagingPlatform) RecordBudgetRule(user bot_interface.BotRecipient, amount float32) ([]bot_interface.Message, error) {
	firstOfMonth, nextMonth := monthInterval()
	currentState, errGettingState := env.Storage.GetUserState(user.UserID)
	if errGettingState != nil {
		log.Print(fmt.Errorf("error getting user state in RecordBudgetRule: %v", errGettingState))
		return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, errGettingState
	}
	selectedTag := trimStringFromFirstSpace(currentState)
	err := env.Storage.CreateTarget(selectedTag, amount, firstOfMonth, nextMonth, user.UserID)
	if err != nil {
		log.Print(fmt.Errorf("error creating target in RecordBudgetRule: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}, errGettingState
	}
	secondMessage, errGettingSecond := env.GiveInstructionOnBudgeting(user)
	if errGettingSecond != nil {
		log.Print(fmt.Errorf("error giving instruction on budgeting in RecordBudgetRule: %v", err))
	}
	return append([]bot_interface.Message{{Text: fmt.Sprintf("Recorded: budget for '%s' is %.2f", selectedTag, amount)}}, secondMessage...), nil
}

func (env MessagingPlatform) GiveCurrentStatistics(user bot_interface.BotRecipient) ([]bot_interface.Message, error) {
	firstOfMonth, nextMonth := monthInterval()
	spending, err := env.Storage.GetMoneyEventsByDateInterval(firstOfMonth, nextMonth, user.UserID)
	if err != nil {
		log.Print(fmt.Errorf("error getting money events in GiveCurrentStatistics: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile. Please try again later"}}, err
	}
	spendingSums := make(map[string]float32)
	for _, spendingItem := range spending {
		spendingSums[spendingItem.Tag] += spendingItem.Amount
	}

	targetSums := make(map[string]float32)
	targets, err := env.Storage.GetTargets(firstOfMonth, nextMonth, user.UserID)
	if err == nil {
		for _, item := range targets {
			targetSums[item.Tag] = item.Amount
		}
	} else {
		log.Print(fmt.Errorf("error getting targets in GiveCurrentStatistics: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile. Please try again later"}}, err
	}
	var resultTags []string
	for key, value := range spendingSums {
		if _, ok := targetSums[key]; ok {
			if targetSums[key] < value {
				resultTags = append(resultTags, fmt.Sprintf("%s: %.2f > %.2f !Warning!", key, value, targetSums[key]))
			} else {
				resultTags = append(resultTags, fmt.Sprintf("%s: %.2f <= %.2f OK", key, value, targetSums[key]))
			}
		} else {
			resultTags = append(resultTags, fmt.Sprintf("%s: %.2f", key, value))
		}
	}
	return []bot_interface.Message{{Text: strings.Join(resultTags, "\n")}, provideMainOptions()}, nil
}

func (env MessagingPlatform) SetSpending(user bot_interface.BotRecipient, amount float32) ([]bot_interface.Message, error) {
	err := env.Storage.SetState(user.UserID, fmt.Sprintf("%s %.2f", bot_interface.StateSpending, amount))
	if err != nil {
		log.Print(fmt.Errorf("error setting user state in SetSpending: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile. Please try again later"}}, err
	}
	acceptedTags, err := env.Storage.GetUserTags(user.UserID)
	if err != nil {
		log.Print(fmt.Errorf("error getting tags in SetSpending: %v", err))
	}
	if len(acceptedTags) > 0 {
		var options []bot_interface.Option
		for _, acceptedTag := range acceptedTags {
			options = append(options, bot_interface.Option{Text: acceptedTag, Id: "inline_" + acceptedTag})
		}
		return []bot_interface.Message{{Text: "For which category do I have to record this expense?", Options: options}}, nil
	} else {
		return []bot_interface.Message{{Text: "For which category do I have to record this expense?", Options: defaultTags}}, nil
	}
}

func (env MessagingPlatform) SetSpendingWithTag(user bot_interface.BotRecipient, amount float32, tag string, comment string) ([]bot_interface.Message, error) {
	err := env.Storage.CreateMoneyEvent(amount, "ARS", comment, tag, user.UserID)
	if err != nil {
		log.Print(fmt.Errorf("error creating money event in SetSpendingWithTag: %v", err))
		return []bot_interface.Message{{Text: "Problem working with your profile. Please try again later"}}, err
	}
	err = env.Storage.SetState(user.UserID, "")
	if err != nil {
		log.Print(fmt.Errorf("error updating state in SetSpendingWithTag: %v", err))
	}
	return []bot_interface.Message{{Text: fmt.Sprintf("Your expense is recorded:\n%.2f - %s", amount, tag)}, provideMainOptions()}, nil
}
