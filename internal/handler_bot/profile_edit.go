package handler_bot

import (
	"fmt"
	"strconv"

	"date-app/assets/localization"
	"date-app/configs"
	"date-app/internal/handler_bot/bot_client"
	"date-app/internal/hash"
	"date-app/internal/profile"
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
	bot *tgbotapi.BotAPI, chatID int64, text string, username string,
) error {
	const op = "HandlerNameChoice"

	UserProfile[chatID] = profile.Profile{Name: text, URL: username}

	UserState[chatID] = StateProfileSexChoice

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
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerSexChoice"

	userProfile := UserProfile[chatID]
	switch text {
	case localization.Russian.ProfileQuestions.Man:
		userProfile.Sex = true
	case localization.Russian.ProfileQuestions.NotMan:
		userProfile.Sex = false
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
	UserProfile[chatID] = userProfile

	UserState[chatID] = StateProfileAgeChoice

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

	userProfile := UserProfile[chatID]
	userProfile.Birthday = timestamp.ToTimestamp(text)
	if userProfile.Birthday == "" {
		err := Send(
			bot, chatID, RemoveKeyboard,
			localization.Russian.ProfileQuestions.IncorrectAge,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	UserProfile[chatID] = userProfile

	UserState[chatID] = StateProfileText
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

	userProfile := UserProfile[chatID]
	userProfile.ProfileText = text
	UserProfile[chatID] = userProfile

	UserState[chatID] = StateProfilePhoto
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
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerEndPhoto"

	// for every text

	userProfile := UserProfile[chatID]
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

	UserState[chatID] = StateWait

	client, isNotFirstProfile := UserClient[chatID]
	if !isNotFirstProfile {
		login := loginRule(chatID)
		password := passwordRule(chatID)
		var ID int
		var err error
		client, ID, err = bot_client.CreateUser(login, password)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		UserClient[chatID] = client
		TelegramUserID[ID] = chatID
	}

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
