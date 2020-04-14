package main

import (
	"./telegram-bot-api"
	"./vozHelpers"
	"strconv"
	s "strings"
)

const Keyword = "#phatcard"
const BotToken = "1058843020:AAH-iNx3A-4O_PGo6B2YTgJv5R4OTzidplg"

var clickRecords []string // Record clickers (string: MessageID + UserID)

func main() {
	bot, _ := tgbotapi.NewBotAPI(BotToken)
	bot.Debug = true

	var u = tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var updates, _ = bot.GetUpdatesChan(u)

	for update := range updates {
		//----------------------------
		// Process #phatcard messages
		//----------------------------
		if update.Message != nil && s.HasPrefix(update.Message.Text, Keyword) {
			var message = update.Message
			var sponsor = message.From.FirstName + " " + message.From.LastName + " @" + message.From.UserName

			// Create inline keyboard buttons
			var cardList = vozHelpers.CreateCardList(message.Text)
			var buttonList = createButtonList(cardList)

			// Send message with created buttons
			var msg = tgbotapi.NewMessage(message.Chat.ID, "Sponsor: "+sponsor)
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup2(buttonList)
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

			// Delete original messages
			var _, err = bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
			if err != nil {
				var errMsg = tgbotapi.NewMessage(message.Chat.ID, "Please give me Delete messages permission!")
				bot.Send(errMsg)
			}
		}

		//----------------------------
		// Process callback queries
		//----------------------------
		if update.CallbackQuery != nil {
			var callback = update.CallbackQuery

			switch {
			// Recall card info if user took the card before
			case s.HasSuffix(callback.Data, "|"+strconv.Itoa(callback.From.ID)):
				// DONE - Alert: cardInfo
				var cardInfo = s.Split(callback.Data, "|")[0]
				sendAlert(bot, callback, cardInfo)

			// User took another card
			case userTookACard(callback.Message.MessageID, callback.From.ID):
				// DONE - Alert: Each member can take maximum 01 card!
				sendAlert(bot, callback, "Each member can take maximum 01 card!")

			// Card is available (and user has NOT taken any card)
			case !s.Contains(callback.Data, "|"):
				// DONE - Edit message with clicked button
				var buttonIndex, _ = strconv.Atoi(s.Split(callback.Data, "-")[0])
				var buttonList = callback.Message.ReplyMarkup
				editButtonList(buttonList, buttonIndex, callback)
				var edit = tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, buttonList)
				bot.Send(edit)

				// DONE - Add to clickerRecords
				recordClicker(callback)

				// DONE - Alert: cardInfo
				sendAlert(bot, callback, callback.Data)

			// Card is already taken by another user
			case s.Contains(callback.Data, "|"):
				// DONE - Alert: This card is already taken by another member!
				sendAlert(bot, callback, "This card is already taken by another member!")
			}
		}
	}
}

// Create a list of buttons (1 button for 1 card) for inline keyboard markup.
func createButtonList(cardList []vozHelpers.Card) [][]tgbotapi.InlineKeyboardButton {
	var buttonList [][]tgbotapi.InlineKeyboardButton
	for i, oneCard := range cardList {
		var button = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(oneCard.Provider, strconv.Itoa(i)+"-"+oneCard.CardInfo))
		buttonList = append(buttonList, button)
	}
	return buttonList
}

// Edit the button if a user click on it.
func editButtonList(inlineKeyboard *tgbotapi.InlineKeyboardMarkup, buttonIndex int, callback *tgbotapi.CallbackQuery) {
	var text = callback.From.FirstName + " " + callback.From.LastName + " @" + callback.From.UserName
	var data = callback.Data + "|" + strconv.Itoa(callback.From.ID)
	var clickedButton = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(text, data))
	inlineKeyboard.InlineKeyboard[buttonIndex] = clickedButton
}

// If a user gets a card, they will be recorded (string of ChatID + MessageID + UserID)
// to avoid getting more than 1 card each time.
func recordClicker(callback *tgbotapi.CallbackQuery) {
	tmp := strconv.Itoa(callback.Message.MessageID) + strconv.Itoa(callback.From.ID)
	clickRecords = append(clickRecords, tmp)
}

func userTookACard(messageId, userId int) bool {
	var tmp = strconv.Itoa(messageId) + strconv.Itoa(userId)
	for _, r := range clickRecords {
		if tmp == r {
			return true
		}
	}
	return false
}

func sendAlert(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, alertMessage string) {
	bot.AnswerCallbackQuery(
		tgbotapi.CallbackConfig{
			CallbackQueryID: callback.ID,
			Text:            alertMessage,
			ShowAlert:       true})
}
