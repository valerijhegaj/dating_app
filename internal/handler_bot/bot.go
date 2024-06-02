package handler_bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"date-app/assets/localization"
	"date-app/configs"
	"date-app/internal/handler_bot/state"
	"date-app/internal/hash"
	"date-app/internal/profile"
	"date-app/internal/timestamp"
	"date-app/internal/token"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var URL = "http://" + configs.Config.TgBot.Host + ":" + strconv.Itoa(configs.Config.Main.Port)

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

var removeKeyboard = tgbotapi.NewRemoveKeyboard(true)

func HandlerOnStart(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	const op = "HandlerOnStart"
	chatID := update.Message.Chat.ID
	state.State[chatID] = state.State2
	err := Send(
		bot, chatID, removeKeyboard,
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

func HandlerOnWait(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	const op = "HandlerOnWait"
	chatID := update.Message.Chat.ID
	_, ok := state.UserID[chatID]
	if !ok {
		err := HandlerOnStart(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	state.State[chatID] = state.State7
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

func HandlerOnFind(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	const op = "HandlerOnFind"
	chatID := update.Message.Chat.ID
	client, ok := state.UserID[chatID]
	if !ok {
		err := HandlerOnStart(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	state.State[chatID] = state.State8
	err := Send(
		bot, chatID, likeScreenKeyboard,
		localization.Russian.LikeScreenMessage,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = ShowIndexed(client, bot, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func ShowIndexed(
	client http.Client, bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	const op = "HandlerOnIndexed"
	r, err := client.Get(URL + "/api/v1/indexed")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	chatID := update.Message.Chat.ID
	if r.StatusCode == http.StatusForbidden {
		err = Send(
			bot, chatID, nil, localization.Russian.NoIndexedLeft,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = HandlerOnWait(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		_ = r.Body.Close()
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	var x struct {
		UserID int `json:"user_id"`
	}
	err = json.Unmarshal(body, &x)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	state.IndexedID[chatID] = x.UserID
	r, err = client.Get(URL + "/api/v1/profile/" + strconv.Itoa(x.UserID))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	body, err = io.ReadAll(r.Body)
	if err != nil {
		_ = r.Body.Close()
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	var p profile.Profile
	err = json.Unmarshal(body, &p)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = ShowProfile(bot, p, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerOnPhoto(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	chatID := update.Message.Chat.ID
	if state.State[chatID] == state.State6 {
		photo := update.Message.Photo[len(update.Message.Photo)-1]
		p := state.Data[chatID]
		p.Photo = append(p.Photo, photo.FileID)
		state.Data[chatID] = p
	}
	return nil
}

var finishProfileKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.Finish),
	),
)

var sexKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.Man),
		tgbotapi.NewKeyboardButton(localization.Russian.ProfileQuestions.NotMan),
	),
)

func HandlerQ2(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ2"
	chatID := update.Message.Chat.ID
	state.Data[chatID] = profile.Profile{Name: update.Message.Text}
	state.State[chatID] = state.State3
	err := Send(
		bot, chatID, sexKeyboard,
		localization.Russian.ProfileQuestions.Sex,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerQ3(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ3"
	chatID := update.Message.Chat.ID
	p := state.Data[chatID]
	if update.Message.Text == localization.Russian.ProfileQuestions.Man {
		p.Sex = true
	} else if update.Message.Text == localization.Russian.ProfileQuestions.NotMan {
		p.Sex = false
	} else {
		err := Send(
			bot, chatID, nil,
			localization.Russian.ProfileQuestions.IncorrectSex,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	state.Data[chatID] = p
	state.State[chatID] = state.State4
	err := Send(
		bot, chatID, removeKeyboard,
		localization.Russian.ProfileQuestions.Age,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerQ4(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ4"
	chatID := update.Message.Chat.ID
	p := state.Data[chatID]
	p.Birthday = timestamp.ToTimestamp(update.Message.Text)
	if p.Birthday == "" {
		err := Send(
			bot, chatID, removeKeyboard,
			localization.Russian.ProfileQuestions.IncorrectAge,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	state.Data[chatID] = p
	state.State[chatID] = state.State5
	err := Send(
		bot, chatID, removeKeyboard,
		localization.Russian.ProfileQuestions.ProfileText,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerQ5(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ5"
	chatID := update.Message.Chat.ID
	p := state.Data[chatID]
	p.ProfileText = update.Message.Text
	state.Data[chatID] = p
	state.State[chatID] = state.State6
	err := Send(
		bot, chatID, finishProfileKeyboard,
		localization.Russian.ProfileQuestions.Photo,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerQ6(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ6"
	chatID := update.Message.Chat.ID
	p := state.Data[chatID]
	if len(p.Photo) == 0 {
		err := Send(
			bot, chatID, nil,
			localization.Russian.ProfileQuestions.NoPhoto,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	state.State[chatID] = state.State7

	var err error
	client := http.Client{}
	client.Jar, err = cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	body := fmt.Sprintf(
		`{"login":"%s", "password":"%s"}`,
		"tg"+strconv.FormatInt(chatID, 16),
		hash.Calculate(
			"tg"+strconv.FormatInt(
				chatID, 16,
			)+configs.Config.Main.HashKey,
		),
	)
	r, err := client.Post(
		URL+"/api/v1/user", "", bytes.NewReader([]byte(body)),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	r, err = client.Post(
		URL+"/api/v1/session", "", bytes.NewReader([]byte(body)),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	u, err := url.Parse(URL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	client.Jar.SetCookies(u, r.Cookies())
	_, ID, err := token.GetFromCookie(r.Cookies()[0])
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	state.UserID[chatID] = client
	profileData, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	r, err = client.Post(
		URL+"/api/v1/profile/"+strconv.Itoa(ID), "",
		bytes.NewReader(profileData),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = ShowProfile(bot, p, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = HandlerOnWait(bot, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerQ7(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ7"
	if update.Message.Text == localization.Russian.WaitScreenKeyboard.Find {
		err := HandlerOnFind(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func HandlerQ8(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	const op = "HandlerQ8"
	chatID := update.Message.Chat.ID
	client := state.UserID[chatID]
	indexedID := strconv.Itoa(state.IndexedID[chatID])
	switch update.Message.Text {
	case localization.Russian.LikeScreenKeyboard.Like:
		r, err := client.Post(
			URL+"/api/v1/like/"+indexedID+"?is_like=1", "",
			bytes.NewReader(nil),
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case localization.Russian.LikeScreenKeyboard.Wait:
		err := HandlerOnWait(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	case localization.Russian.LikeScreenKeyboard.Dislike:
		r, err := client.Post(
			URL+"/api/v1/like/"+indexedID+"?is_like=0", "",
			bytes.NewReader(nil),
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case localization.Russian.LikeScreenKeyboard.Report:
		// not implemented
		r, err := client.Post(
			URL+"/api/v1/like/"+indexedID+"?is_like=0", "",
			bytes.NewReader(nil),
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	err := ShowIndexed(client, bot, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func HandlerOnText(
	bot *tgbotapi.BotAPI, update tgbotapi.Update,
) error {
	const op = "HandlerOnText"
	chatID := update.Message.Chat.ID
	q, ok := state.State[chatID]
	if !ok {
		err := HandlerOnStart(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	switch q {
	case state.State2:
		err := HandlerQ2(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State3:
		err := HandlerQ3(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State4:
		err := HandlerQ4(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State5:
		err := HandlerQ5(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State6:
		err := HandlerQ6(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State7:
		err := HandlerQ7(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case state.State8:
		err := HandlerQ8(bot, update)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
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
