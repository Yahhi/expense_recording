package speaking

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/tucnak/telebot.v2"
	"ingresos_gastos/storage_interface"
)

const (
	StateCreateTags     = "tag_create"
	StateModifyBudget   = "tag_budget"
	StateSpending       = "tag_spending"
	CommandCancel       = "cancel"
	CommandStart        = "start"
	CommandHelp         = "help"
	CommandDefineTags   = "define_tags"
	CommandDefineBudget = "define_budget"
	CommandStatistics   = "view_statistics"
	//CommandFeedback     = "feedback"
)

type Env struct {
	Storage storage_interface.ActualStorage
	Bot     *telebot.Bot
}

func (env Env) ProvideGreeting(user *telebot.User) {
	reply := "Hello, I'm your budget bot! I can help you manage your expenses and manage budget."
	_, err := env.Bot.Send(user, reply)
	if err != nil {
		log.Print(fmt.Errorf("error sending message in ProvideGreeting: %v", err))
		return
	}
	exists, errGettingUser := env.Storage.UserExists(user.ID)
	if errGettingUser != nil {
		log.Print(fmt.Errorf("error checking user existance in ProvideGreeting: %v", err))
	}
	if !exists {
		errInDb := env.Storage.CreateUser(user.ID, user.Username)
		if errInDb != nil {
			log.Print(fmt.Errorf("error creating user in ProvideGreeting: %v", errInDb))
			return
		}
	}
	env.ProvideMainOptions(user)
}

func (env Env) ProvideHelp(user *telebot.User) {
	reply := fmt.Sprintf(`%s - Start the bot and get a description
%s - Get a list of all available commands
%s - Define categories of expenses
%s - Set a budget for each category for the current month
<number> <tag> <comment> - save a new expense (only number is required, other fields are optional)
%s - View your current month statistics
%s - finish ongoing operation (like defining month budget or creating tags)`, CommandStart, CommandHelp, CommandDefineTags, CommandDefineBudget, CommandStatistics, CommandCancel)
	_, err := env.Bot.Send(user, reply)
	if err != nil {
		log.Print(fmt.Errorf("error sending message in ProvideHelp: %v", err))
		return
	}
}

func (env Env) ProvideMainOptions(user *telebot.User) {
	text := "To add new expense just type it here. Other commands:"
	options := map[string]string{
		CommandHelp:         "\xE2\x9D\x93help",
		CommandDefineTags:   "\xE2\x9C\x8Ftags",
		CommandDefineBudget: "\xF0\x9F\x92\xB0budget",
		CommandStatistics:   "\xF0\x9F\x93\x8Astatistics",
	}
	_, err := env.Bot.Send(user, text, env.convertMapToOptions(options))
	if err != nil {
		log.Print(fmt.Errorf("error sending message in ProvideMainOptions: %v", err))
		return
	}
}

func (env Env) GiveInstructionsOnTags(user *telebot.User) {
	acceptedTags, err := env.Storage.GetUserTags(user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting user tags in GiveInstructionsOnTags: %v", err))
		return
	}
	var replyOptions []string
	var textReply string
	if len(acceptedTags) == 0 {
		replyOptions = defaultTags
		textReply = "You didn't select tags yet. Please select from the list to add or input your own by keyboard"
	} else {
		replyOptions = acceptedTags
		textReply = "Your current tags are below.\nSelect a tag to delete or input new tags by keyboard"
	}
	options := env.convertStringsToOptions(replyOptions, true)
	errSavingState := env.Storage.SetState(user.ID, StateCreateTags)
	if errSavingState != nil {
		log.Print(fmt.Errorf("error saving user state in GiveInstructionsOnTags: %v", errSavingState))
		return
	}
	_, errSending := env.Bot.Send(user, textReply, options)
	if errSending != nil {
		log.Print(fmt.Errorf("error sending message in GiveInstructionsOnTags: %v", errSending))
		return
	}
}

func (env Env) CancelLastState(user *telebot.User) {
	errSavingState := env.Storage.SetState(user.ID, "")
	if errSavingState != nil {
		log.Print(fmt.Errorf("error saving user state in GiveInstructionsOnTags: %v", errSavingState))
		return
	}
	env.ProvideMainOptions(user)
}

func (env Env) updateTag(user *telebot.User, tag string, alreadyExists bool) {
	if alreadyExists {
		err := env.Storage.RemoveTagForUser(tag, user.ID)
		if err != nil {
			log.Print(fmt.Errorf("error removing tag in updateTag: %v", err))
			return
		}
	} else {
		err := env.Storage.AddTagForUser(tag, user.ID)
		if err != nil {
			log.Print(fmt.Errorf("error adding tag in updateTag: %v", err))
			return
		}
	}
}

func (env Env) UpdateTags(user *telebot.User, tags []string) {
	existingTags, err := env.Storage.GetUserTags(user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting user tags in UpdateTags: %v", err))
	}
	for _, item := range tags {
		exists := slices.Contains(existingTags, item)
		env.updateTag(user, item, exists)
	}
	env.GiveInstructionsOnTags(user)
}

func (env Env) GiveInstructionOnBudgeting(user *telebot.User) {
	firstOfMonth, nextMonth := env.monthInterval()
	existingData, err := env.Storage.GetTargets(firstOfMonth, nextMonth, user.ID)
	replyOptions := make(map[string]string)
	textReply := "To update budget for this month select a tag and then enter target amount"
	if err == nil {
		for _, existingData := range existingData {
			replyOptions[existingData.Tag] = fmt.Sprintf("%s - %.2f", existingData.Tag, existingData.Amount)
		}
	} else {
		log.Print(fmt.Errorf("error getting targets in GiveInstructionOnBudgeting: %v", err))
	}
	otherTags, errGettingTags := env.Storage.GetUserTags(user.ID)
	if errGettingTags == nil {
		for _, tag := range otherTags {
			if _, ok := replyOptions[tag]; !ok {
				replyOptions[tag] = tag
			}
		}
	} else {
		log.Print(fmt.Errorf("error getting tags in GiveInstructionOnBudgeting: %v", errGettingTags))
	}

	errSettingState := env.Storage.SetState(user.ID, StateModifyBudget)
	if errSettingState != nil {
		log.Print(fmt.Errorf("error setting user state in GiveInstructionOnBudgeting: %v", errSettingState))
		return
	}
	_, errSending := env.Bot.Send(user, textReply, env.convertMapToOptions(replyOptions))
	if errSending != nil {
		log.Print(fmt.Errorf("error sending message in GiveInstructionOnBudgeting: %v", errSending))
		return
	}
}

func (env Env) ConfirmSelectingBudgetTag(user *telebot.User, tag string) {
	_, errSending := env.Bot.Send(user, "Enter updated amount of money you want to spend in this month (or 0 if you don't want to spend money for this)")
	if errSending != nil {
		log.Print(fmt.Errorf("error sending message in ConfirmSelectingBudgetTag: %v", errSending))
		return
	}
	err := env.Storage.SetState(user.ID, fmt.Sprintf("%s %s", StateModifyBudget, tag))
	if err != nil {
		log.Print(fmt.Errorf("error setting state in ConfirmSelectingBudgetTag: %v", err))
		return
	}
}

func (env Env) RecordBudgetRule(user *telebot.User, amount float32) {
	firstOfMonth, nextMonth := env.monthInterval()
	currentState, errGettingState := env.Storage.GetUserState(user.ID)
	if errGettingState != nil {
		log.Print(fmt.Errorf("error getting user state in RecordBudgetRule: %v", errGettingState))
		return
	}
	selectedTag := trimStringFromFirstSpace(currentState)
	err := env.Storage.CreateTarget(selectedTag, amount, firstOfMonth, nextMonth, user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error creating target in RecordBudgetRule: %v", err))
		return
	}
	env.GiveInstructionOnBudgeting(user)
}

func (env Env) SetSpending(user *telebot.User, amount float32) {
	err := env.Storage.SetState(user.ID, fmt.Sprintf("%s %.2f", StateSpending, amount))
	if err != nil {
		log.Print(fmt.Errorf("error setting user state in SetSpending: %v", err))
		return
	}
	acceptedTags, err := env.Storage.GetUserTags(user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting tags in SetSpending: %v", err))
		return
	}
	if len(acceptedTags) > 0 {
		_, errSending := env.Bot.Send(user, "For which category do I have to record this expense?", env.convertStringsToOptions(acceptedTags, false))
		if errSending != nil {
			log.Print(fmt.Errorf("error sending message in SetSpending: %v", errSending))
			return
		}
	} else {
		err := env.Storage.SetState(user.ID, "")
		if err != nil {
			log.Print(fmt.Errorf("error SetState: %v", err))
			return
		}
		_, errSending := env.Bot.Send(user, "First you need to create tags")
		if errSending != nil {
			log.Print(fmt.Errorf("error sending no message in SetSpending: %v", errSending))
			return
		}
	}
}

func (env Env) SetSpendingWithTag(user *telebot.User, amount float32, tag string, comment string) {
	err := env.Storage.CreateMoneyEvent(amount, "ARS", comment, tag, user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error creating money event in SetSpendingWithTag: %v", err))
		return
	}
	err = env.Storage.SetState(user.ID, "")
	if err != nil {
		log.Print(fmt.Errorf("error updating state in SetSpendingWithTag: %v", err))
		return
	}
	_, errSending := env.Bot.Send(user, fmt.Sprintf("Your expense is recorded:\n%.2f - %s", amount, tag))
	if errSending != nil {
		log.Print(fmt.Errorf("error sending message in SetSpendingWithTag: %v", errSending))
		return
	}
}

func (env Env) GiveCurrentStatistics(user *telebot.User) {
	firstOfMonth, nextMonth := env.monthInterval()
	spending, err := env.Storage.GetMoneyEventsByDateInterval(firstOfMonth, nextMonth, user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting money events in GiveCurrentStatistics: %v", err))
		return
	}
	spendingSums := make(map[string]float32)
	for _, spendingItem := range spending {
		spendingSums[spendingItem.Tag] += spendingItem.Amount
	}

	targetSums := make(map[string]float32)
	targets, err := env.Storage.GetTargets(firstOfMonth, nextMonth, user.ID)
	if err == nil {
		for _, item := range targets {
			targetSums[item.Tag] = item.Amount
		}
	} else {
		log.Print(fmt.Errorf("error getting targets in GiveCurrentStatistics: %v", err))
		return
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
	_, errSending := env.Bot.Send(user, strings.Join(resultTags, "\n"))
	if errSending != nil {
		log.Print(fmt.Errorf("error sending message in GiveCurrentStatistics: %v", errSending))
		return
	}
}

func (env Env) DetectAppropriateActionForButton(user *telebot.User, inlineButtonTag string) {
	switch inlineButtonTag {
	case CommandDefineTags:
		env.GiveInstructionsOnTags(user)
	case CommandDefineBudget:
		env.GiveInstructionOnBudgeting(user)
	case CommandStatistics:
		env.GiveCurrentStatistics(user)
	case CommandHelp:
		env.ProvideHelp(user)
	case CommandStart:
		env.ProvideGreeting(user)
	case CommandCancel:
		env.CancelLastState(user)
	}

	userState, err := env.Storage.GetUserState(user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting user state in GiveCurrentStatistics: %v", err))
		return
	}

	switch userState {
	case StateCreateTags:
		env.UpdateTags(user, []string{strings.TrimPrefix(inlineButtonTag, "inline_")})
	case StateModifyBudget:
		env.ConfirmSelectingBudgetTag(user, strings.TrimPrefix(inlineButtonTag, "inline_"))
	case StateSpending:
		numberToParse := trimStringFromFirstSpace(userState)
		amount, errParsing := strconv.ParseFloat(numberToParse, 32)
		if errParsing != nil {
			log.Print(fmt.Errorf("error parsing float from state '%s' in GiveCurrentStatistics: %v", userState, err))
			return
		}
		env.SetSpendingWithTag(user, float32(amount), strings.TrimPrefix(inlineButtonTag, "inline_"), "")
	default:
		//unrecognized. Let's write an error
		log.Print("ERROR unrecognized button: " + inlineButtonTag)
	}
}

func (env Env) DetectAppropriateActionForInput(user *telebot.User, messageText string) {
	userState, err := env.Storage.GetUserState(user.ID)
	if err != nil {
		log.Print(fmt.Errorf("error getting user state in DetectAppropriateActionForInput: %v", err))
		return
	}
	if strings.HasPrefix(userState, StateCreateTags) {
		if strings.Contains(messageText, " ") {
			tags := strings.Split(messageText, " ")
			env.UpdateTags(user, tags)
		} else {
			env.UpdateTags(user, []string{messageText})
		}
	} else if strings.HasPrefix(userState, StateModifyBudget) {
		amount, err := strconv.ParseFloat(messageText, 32)
		if err == nil {
			env.RecordBudgetRule(user, float32(amount))
		} else {
			log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
			_, err := env.Bot.Send(user, "I can't understand your number, please enter correct number")
			if err != nil {
				log.Print(fmt.Errorf("error sending message in DetectAppropriateActionForInput: %v", err))
				return
			}
		}
	} else {
		if strings.Contains(messageText, " ") {
			parts := strings.Split(messageText, " ")
			possibleAmount, err := strconv.ParseFloat(parts[0], 32)
			if err == nil {
				contentLength := len(parts)
				if contentLength == 3 { // todo switch
					env.SetSpendingWithTag(user, float32(possibleAmount), parts[1], parts[2])
				} else if contentLength == 2 {
					env.SetSpendingWithTag(user, float32(possibleAmount), parts[1], "")
				} else {
					env.SetSpending(user, float32(possibleAmount))
				}
			} else {
				log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
				err := env.Storage.SaveFeedback(user.ID, messageText)
				if err != nil {
					log.Print(fmt.Errorf("error saving feedback in DetectAppropriateActionForInput: %v", err))
					return
				}
				_, errSending := env.Bot.Send(user, "We recognize your input as feedback. Saved it for our developers")
				if errSending != nil {
					log.Print(fmt.Errorf("error sending message in DetectAppropriateActionForInput: %v", errSending))
					return
				}
			}
		} else {
			possibleAmount, err := strconv.ParseFloat(messageText, 32)
			if err == nil {
				env.SetSpending(user, float32(possibleAmount))
			} else {
				log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
				_, err := env.Bot.Send(user, "I can't understand your number, please enter correct number")
				if err != nil {
					log.Print(fmt.Errorf("error sending message in DetectAppropriateActionForInput: %v", err))
					return
				}
			}
		}
	}
}
