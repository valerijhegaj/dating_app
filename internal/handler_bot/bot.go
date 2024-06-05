package handler_bot

import (
	"fmt"
	"log"
	"net/http"

	"date-app/assets/localization"
	"date-app/internal/handler_bot/bot_client"
	"date-app/internal/profile"
	"date-app/internal/timestamp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Send(
	bot *tgbotapi.BotAPI, chatID int64, replyMarkup interface{},
	text string,
) error {
	const op = "Send"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = replyMarkup
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

var RemoveKeyboard = tgbotapi.NewRemoveKeyboard(true)

func HandlerOnStart(bot *tgbotapi.BotAPI, chatID int64) error {
	const op = "HandlerOnStart"

	UserState[chatID] = StateProfileNameChoice
	err := Send(
		bot, chatID, RemoveKeyboard,
		localization.Russian.StartMessage,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

var waitScreenKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.WaitScreenKeyboard.Find),
		tgbotapi.NewKeyboardButton(localization.Russian.WaitScreenKeyboard.ShowProfile),
		tgbotapi.NewKeyboardButton(localization.Russian.WaitScreenKeyboard.ChangeProfile),
	),
)

func HandlerOnWait(bot *tgbotapi.BotAPI, chatID int64) error {
	const op = "HandlerOnWait"

	_, ok := UserClient[chatID]
	if !ok {
		err := HandlerOnStart(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	UserState[chatID] = StateWait

	err := Send(
		bot, chatID, waitScreenKeyboard,
		localization.Russian.WaitMessage,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

var likeScreenKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.LikeScreenKeyboard.Like),
		tgbotapi.NewKeyboardButton(localization.Russian.LikeScreenKeyboard.Dislike),
		tgbotapi.NewKeyboardButton(localization.Russian.LikeScreenKeyboard.Report),
		tgbotapi.NewKeyboardButton(localization.Russian.LikeScreenKeyboard.Wait),
	),
)

func HandlerOnFind(bot *tgbotapi.BotAPI, chatID int64) error {
	const op = "HandlerOnFind"

	client, ok := UserClient[chatID]
	if !ok {
		err := HandlerOnStart(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	UserState[chatID] = StateLike

	err := Send(
		bot, chatID, likeScreenKeyboard,
		localization.Russian.LikeScreenMessage,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = ShowIndexed(client, bot, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func ShowIndexed(
	client http.Client, bot *tgbotapi.BotAPI, chatID int64,
) error {
	const op = "HandlerOnIndexed"

	indexedID, err := bot_client.GetIndexed(client)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if indexedID == 0 {
		err = Send(
			bot, chatID, nil, localization.Russian.NoIndexedLeft,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = HandlerOnWait(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	IndexedID[chatID] = indexedID

	p, err := bot_client.GetProfile(client, indexedID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = ShowProfile(bot, p, chatID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerOnPhoto(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	chatID := update.Message.Chat.ID
	state, ok := UserState[chatID]
	if ok && state == StateProfilePhoto {
		photo := update.Message.Photo[len(update.Message.Photo)-1]
		p := UserProfile[chatID]
		p.Photo = append(p.Photo, photo.FileID)
		UserProfile[chatID] = p
	}
	return nil
}

func HandlerTextStateWait(
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerTextStateWait"

	switch text {
	case localization.Russian.WaitScreenKeyboard.Find:
		if err := HandlerOnFind(bot, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case localization.Russian.WaitScreenKeyboard.ChangeProfile:
		if err := HandlerOnStart(bot, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case localization.Russian.WaitScreenKeyboard.ShowProfile:
		userProfile := UserProfile[chatID]
		if err := ShowProfile(bot, userProfile, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func HandlerOnLike(
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerOnLike"

	client := UserClient[chatID]
	likeID := IndexedID[chatID]
	switch text {
	case localization.Russian.LikeScreenKeyboard.Like:
		like, err := bot_client.PostLike(client, likeID, true)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if like.UserID == 0 {
			break
		}

		notificationChatID := TelegramUserID[like.UserID]
		if err = ChangeToStateMatch(bot, notificationChatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err = ChangeToStateMatch(bot, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	case localization.Russian.LikeScreenKeyboard.Wait:
		err := HandlerOnWait(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	case localization.Russian.LikeScreenKeyboard.Dislike:
		_, err := bot_client.PostLike(client, likeID, false)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case localization.Russian.LikeScreenKeyboard.Report:
		// not implemented
		_, err := bot_client.PostLike(client, likeID, false)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func HandlerTextLikeState(
	bot *tgbotapi.BotAPI, chatID int64, text string,
) error {
	const op = "HandlerTextLikeState"

	if err := HandlerOnLike(bot, chatID, text); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	switch UserState[chatID] {
	case StateLike:
		if err := ShowIndexed(
			UserClient[chatID], bot, chatID,
		); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	case StateLikePreMatch:
		UserState[chatID] = StateMatch
		if err := HandlerMatch(bot, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	default:
		return fmt.Errorf("%s: strange state", op)
	}
}

var matchScreenKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.Match.Next),
	),
)

func ChangeToStateMatch(bot *tgbotapi.BotAPI, chatID int64) error {
	const op = "ChangeToStateMatch"

	st := UserState[chatID]
	switch st {
	case StateLike:
		if err := Send(
			bot, chatID, nil, localization.Russian.Match.LikeScreen,
		); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		UserState[chatID] = StateLikePreMatch
	case StateWait:
		if err := Send(
			bot, chatID, matchScreenKeyboard,
			localization.Russian.Match.WaitScreen,
		); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		UserState[chatID] = StateMatch
	default:
	}
	return nil
}

func ShowMatch(bot *tgbotapi.BotAPI, chatID int64) (bool, error) {
	const op = "ShowMatch"
	client := UserClient[chatID]

	likes, err := bot_client.GetLikes(client)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if len(likes) == 0 {
		return false, nil
	}
	for _, like := range likes {
		p, err := bot_client.GetProfile(client, like.UserID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if err = ShowProfile(bot, p, chatID); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if err = Send(
			bot, chatID, matchScreenKeyboard,
			localization.Russian.Match.Messeage+"@"+p.URL,
		); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if err = bot_client.PostProfileViewed(
			client, like.UserID,
		); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	return false, nil
}

func HandlerMatch(bot *tgbotapi.BotAPI, chatID int64) error {
	const op = "HandlerMatch"

	wasMatch, err := ShowMatch(bot, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !wasMatch {
		if err = Send(
			bot, chatID, RemoveKeyboard,
			localization.Russian.Match.FinishMatch,
		); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err = HandlerOnWait(bot, chatID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func HandlerOnText(
	bot *tgbotapi.BotAPI, chatID int64, msg *tgbotapi.Message,
) error {
	const op = "HandlerOnText"
	text := msg.Text

	q, ok := UserState[chatID]
	if !ok {
		err := HandlerOnStart(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	switch q {
	case StateProfileNameChoice:
		err := HandlerNameChoice(bot, chatID, text, msg.Chat.UserName)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateProfileSexChoice:
		err := HandlerSexChoice(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateProfileAgeChoice:
		err := HandlerAgeChoice(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateProfileText:
		err := HandlerProfileText(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateProfilePhoto:
		err := HandlerEndPhoto(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateWait:
		err := HandlerTextStateWait(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateLike:
		err := HandlerTextLikeState(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateLikePreMatch:
		err := HandlerTextLikeState(bot, chatID, text)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateMatch:
		err := HandlerMatch(bot, chatID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case StateNonAuthed:
	default:
		log.Printf("strage state %q", q)
	}
	return nil
}

func ShowProfile(
	bot *tgbotapi.BotAPI, p profile.Profile, chatID int64,
) error {
	const op = "ShowProfile"
	var photos []interface{}

	for i, ph := range p.Photo {
		x := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(ph))
		if i == 0 {
			x.Caption = p.Name + ", " + timestamp.ToAge(p.Birthday) + ", " + p.ProfileText
		}
		photos = append(photos, x)
	}
	m := tgbotapi.NewMediaGroup(chatID, photos)
	_, err := bot.SendMediaGroup(m)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
