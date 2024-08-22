package handler_bot

import (
	"fmt"
	"strconv"

	"date-app/assets/localization"
	"date-app/configs"
	"date-app/internal/handler_bot/bot_client"
	"date-app/internal/hash"
	"date-app/internal/timestamp"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var sexChoiceKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.Man),
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.NotMan),
	),
)

var finishProfileKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.Finish),
	),
)

func HandlerNameChoice(
	bot *tgbotapi.BotAPI, chatID int64, name string, username string,
) error {
	const op = "HandlerNameChoice"

	Manager.UpdateProfileName(chatID, name)
	Manager.UpdateProfileURL(chatID, username)

	Manager.UpdateState(chatID, StateProfileSexChoice)

	err := Send(
		bot, chatID, sexChoiceKeyboard,
		localization.Russian.ProfileQuestions.Sex,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerSexChoice(
	bot *tgbotapi.BotAPI, chatID int64, sexText string,
) error {
	const op = "HandlerSexChoice"

	var sex bool
	switch sexText {
	case localization.Russian.ProfileQuestions.Man:
		sex = true
	case localization.Russian.ProfileQuestions.NotMan:
		sex = false
	default:
		err := Send(
			bot, chatID, nil,
			localization.Russian.ProfileQuestions.IncorrectSex,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	Manager.UpdateProfileSex(chatID, sex)

	Manager.UpdateState(chatID, StateProfileAgeChoice)

	err := Send(
		bot, chatID, RemoveKeyboard,
		localization.Russian.ProfileQuestions.Age,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerAgeChoice(
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerAgeChoice"

	birthday := timestamp.ToTimestamp(text)
	if birthday == "" {
		err := Send(
			bot, chatID, RemoveKeyboard,
			localization.Russian.ProfileQuestions.IncorrectAge,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	Manager.UpdateProfileBirthday(chatID, birthday)

	Manager.UpdateState(chatID, StateProfileText)

	err := Send(
		bot, chatID, RemoveKeyboard,
		localization.Russian.ProfileQuestions.ProfileText,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerProfileText(
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerProfileText"

	Manager.UpdateProfileText(chatID, text)

	Manager.UpdateState(chatID, StateProfilePhoto)

	err := Send(
		bot, chatID, finishProfileKeyboard,
		localization.Russian.ProfileQuestions.Photo,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func loginRule(chatID int64) string {
	return "tg" + strconv.FormatInt(chatID, 16)
}

func passwordRule(chatID int64) string {
	return hash.Calculate(
		"tg" + strconv.FormatInt(
			chatID, 16,
		) + configs.Config.Main.HashKey,
	)
}

func HandlerEndPhoto(
	bot *tgbotapi.BotAPI, chatID int64,
) error {
	const op = "HandlerEndPhoto"

	// for every text

	userProfile := Manager.GetProfile(chatID)
	if len(userProfile.Photo) == 0 {
		err := Send(
			bot, chatID, nil,
			localization.Russian.ProfileQuestions.NoPhoto,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	isFirstProfile := !Manager.CheckClient(chatID)
	if isFirstProfile {
		login := loginRule(chatID)
		password := passwordRule(chatID)
		client, ID, err := bot_client.CreateUser(login, password)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		Manager.UpdateClient(chatID, client)
		Manager.UpdateTgUserID(chatID, ID)
	}
	client := Manager.GetClient(chatID)

	err := bot_client.UpdateProfile(client, userProfile)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = ShowProfile(bot, userProfile, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = HandlerOnWait(bot, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
