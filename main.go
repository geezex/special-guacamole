package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	adminID    = int64(6683327018) // ← замените на ваш Telegram ID
	userStates = make(map[int64]string)
)

func main() {
	log.Println("Starting bot...")

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID

	switch {
	case msg.IsCommand():
		if msg.Command() == "start" {
			userStates[userID] = "waiting_confirm"

			startBtn := tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "confirm")
			keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(startBtn))

			resp := tgbotapi.NewMessage(msg.Chat.ID, "Добро пожаловать! Нажмите кнопку ниже, чтобы подтвердить.")
			resp.ReplyMarkup = keyboard
			bot.Send(resp)
		}

	case userStates[userID] == "waiting_screenshot" && msg.Photo != nil && msg.Caption != "":
		// Пересылаем сообщение администратору
		forward := tgbotapi.NewForward(adminID, msg.Chat.ID, msg.MessageID)
		bot.Send(forward)

	default:
		// Игнорируем все остальные сообщения
		return
	}
}

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	switch query.Data {
	case "confirm":
		userStates[userID] = "waiting_payment"

		payBtn := tgbotapi.NewInlineKeyboardButtonData("Подтвердить оплату", "payment_confirm")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(payBtn))

		edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Подтвердите оплату:")
		edit.ReplyMarkup = &keyboard
		bot.Send(edit)

	case "payment_confirm":
		userStates[userID] = "waiting_screenshot"

		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Отправьте скриншот оплаты с вашими данными для связи в подписи к скриншоту.")
		bot.Send(msg)
	}

	// Удаляем кружочек загрузки на кнопке
	bot.Request(tgbotapi.NewCallback(query.ID, ""))
}
