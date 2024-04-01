package speaking

import (
	"fmt"
	"ingresos_gastos/bot_interface"
	"log"
	"strconv"
	"strings"
)

func (env MessagingPlatform) saveUsageLog(requestType string, userId int64) {
	err := env.Storage.SaveUsageLog(userId, requestType)
	if err != nil {
		log.Print(fmt.Errorf("error saving user log in saveUsageLog: %v", err))
	}
}

func (env MessagingPlatform) DetectAppropriateActionForButton(user bot_interface.BotRecipient, inlineButtonTag string) ([]bot_interface.Message, error) {
	var messages []bot_interface.Message
	var err error
	commandFound := true
	switch inlineButtonTag {
	case bot_interface.CommandDefineTags:
		messages, err = env.GiveInstructionsOnTags(user)
	case bot_interface.CommandDefineBudget:
		messages, err = env.GiveInstructionOnBudgeting(user)
	case bot_interface.CommandStatistics:
		messages, err = env.GiveCurrentStatistics(user)
	case bot_interface.CommandHelp:
		messages, err = env.ProvideHelp(user)
	case bot_interface.CommandStart:
		messages, err = env.ProvideGreeting(user)
	case bot_interface.CommandFeedback:
		messages, err = env.ProvideFeedbackInstruction(user)
	case bot_interface.CommandCancel:
		messages, err = env.CancelLastState(user)
	default:
		commandFound = false
	}
	if commandFound {
		env.saveUsageLog(inlineButtonTag, user.UserID)
		return messages, nil
	}

	userState, err := env.Storage.GetUserState(user.UserID)
	if err == nil {
		switch userState {
		case bot_interface.StateCreateTags:
			messages, err = env.UpdateTag(user, strings.TrimPrefix(inlineButtonTag, "inline_"))
		case bot_interface.StateModifyBudget:
			messages, err = env.ConfirmSelectingBudgetTag(user, strings.TrimPrefix(inlineButtonTag, "inline_"))
		case bot_interface.StateSpending:
			numberToParse := trimStringFromFirstSpace(userState)
			amount, errParsing := strconv.ParseFloat(numberToParse, 32)
			if errParsing == nil {
				messages, err = env.SetSpendingWithTag(user, float32(amount), strings.TrimPrefix(inlineButtonTag, "inline_"), "")
			} else {
				log.Print(fmt.Errorf("error parsing float from state '%s' in GiveCurrentStatistics: %v", userState, err))
				messages = []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}
			}
		default: //unrecognized. Let's write an error
			log.Print("ERROR unrecognized button: " + inlineButtonTag)
			messages = []bot_interface.Message{{Text: "I didn't recognize the command. Sorry"}}
		}
		if err != nil && (messages == nil || len(messages) == 0) {
			messages = []bot_interface.Message{{Text: "Problem in our system. Please try again later"}}
		}
	} else {
		log.Print(fmt.Errorf("error getting user state in DetectAppropriateActionForButton: %v", err))
		messages = []bot_interface.Message{{Text: "Problem reading your profile. Please try again later"}}
	}
	return messages, nil
}

func (env MessagingPlatform) DetectAppropriateActionForInput(user bot_interface.BotRecipient, messageText string) ([]bot_interface.Message, error) {
	var messages []bot_interface.Message
	userState, errGettingState := env.Storage.GetUserState(user.UserID)
	if errGettingState == nil {
		if strings.HasPrefix(userState, bot_interface.StateCreateTags) {
			if strings.Contains(messageText, " ") {
				tags := strings.Split(messageText, " ")
				for _, tag := range tags {
					messageForTag, _ := env.UpdateTag(user, tag)
					messages = append(messages, messageForTag[0])
				}
			} else {
				messages, _ = env.UpdateTag(user, messageText)
			}
		} else if strings.HasPrefix(userState, bot_interface.StateModifyBudget) {
			amount, err := strconv.ParseFloat(messageText, 32)
			if err == nil {
				messages, err = env.RecordBudgetRule(user, float32(amount))
			} else {
				messages = []bot_interface.Message{{Text: "I can't understand your number, please enter correct number"}}
				log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
			}
		} else {
			if strings.Contains(messageText, " ") {
				parts := strings.Split(messageText, " ")
				possibleAmount, err := strconv.ParseFloat(parts[0], 32)
				if err == nil {
					contentLength := len(parts)
					if contentLength == 3 { // todo switch
						messages, err = env.SetSpendingWithTag(user, float32(possibleAmount), parts[1], parts[2])
					} else if contentLength == 2 {
						messages, err = env.SetSpendingWithTag(user, float32(possibleAmount), parts[1], "")
					} else {
						messages, err = env.SetSpending(user, float32(possibleAmount))
					}
				} else {
					log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
					messages, err = env.SaveFeedback(user, messageText)
				}
			} else {
				possibleAmount, err := strconv.ParseFloat(messageText, 32)
				if err == nil {
					messages, err = env.SetSpending(user, float32(possibleAmount))
				} else {
					log.Print(fmt.Errorf("error parsing float from message '%s' in DetectAppropriateActionForInput: %v", messageText, err))
					messages = []bot_interface.Message{{Text: "I can't understand your number, please enter correct number"}}
				}
			}
		}
	} else {
		log.Print(fmt.Errorf("error getting user state in DetectAppropriateActionForInput: %v", errGettingState))
		messages = []bot_interface.Message{{Text: "Problem working with your profile in our system. Please try again later"}}
	}
	return messages, nil
}

func (env MessagingPlatform) ListenToCommands() {
	env.Bot.ListenToCommand("/"+bot_interface.CommandStart, env.ProvideGreeting)
	env.Bot.ListenToCommand("/"+bot_interface.CommandHelp, env.ProvideHelp)
	env.Bot.ListenToCommand("/"+bot_interface.CommandDefineTags, env.GiveInstructionsOnTags)
	env.Bot.ListenToCommand("/"+bot_interface.CommandDefineBudget, env.GiveInstructionOnBudgeting)
	env.Bot.ListenToCommand("/"+bot_interface.CommandStatistics, env.GiveCurrentStatistics)
	env.Bot.ListenToCommand("/"+bot_interface.CommandFeedback, env.ProvideFeedbackInstruction)
	env.Bot.ListenToCommand("/"+bot_interface.CommandCancel, env.CancelLastState)
}

func (env MessagingPlatform) ListenToUserInput() {
	env.Bot.ListenToInput(env.DetectAppropriateActionForInput)
}

func (env MessagingPlatform) ListenToInlineActions() {
	env.Bot.ListenToInlineActions(env.DetectAppropriateActionForButton)
}
