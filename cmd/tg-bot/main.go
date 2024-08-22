package main

import (
	"log"

	"date-app/configs"
	"date-app/internal/handler_bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(configs.Config.TgBot.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			if update.Message.Text != "" {
				text := update.Message.Text
				switch text {
				case "/start":
					err = handler_bot.HandlerOnStart(bot, chatID)
					if err != nil {
						log.Println(err)
					}
				case "/wait":
					err = handler_bot.HandlerOnWait(bot, chatID)
					if err != nil {
						log.Println(err)
					}
				case "/find":
					err = handler_bot.HandlerOnFind(bot, chatID)
					if err != nil {
						log.Println(err)
					}
				default:
					err = handler_bot.HandlerOnText(bot, chatID, update.Message)
					if err != nil {
						log.Println(err)
					}
				}
			} else if update.Message.Photo != nil {
				err = handler_bot.HandlerOnPhoto(bot, update)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}
