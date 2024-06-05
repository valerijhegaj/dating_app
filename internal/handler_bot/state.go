package handler_bot

import (
	"net/http"

	"date-app/internal/profile"
)

var UserState = make(map[int64]int)
var UserProfile = make(map[int64]profile.Profile)
var UserClient = make(map[int64]http.Client)
var IndexedID = make(map[int64]int)
var TelegramUserID = make(map[int]int64)

const (
	StateNonAuthed = iota
	StateProfileNameChoice
	StateProfileSexChoice
	StateProfileAgeChoice
	StateProfileText
	StateProfilePhoto
	StateWait
	StateLike
	StateLikePreMatch
	StateMatch
)
