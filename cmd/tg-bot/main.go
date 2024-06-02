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
			if update.Message.Text != "" {
				switch update.Message.Text {
				case "/start":
					err = handler_bot.HandlerOnStart(bot, update)
					if err != nil {
						log.Println(err)
					}
				case "/wait":
					err = handler_bot.HandlerOnWait(bot, update)
					if err != nil {
						log.Println(err)
					}
				case "/find":
					err = handler_bot.HandlerOnFind(bot, update)
					if err != nil {
						log.Println(err)
					}
				default:
					err = handler_bot.HandlerOnText(bot, update)
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
