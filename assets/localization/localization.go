package localization

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed russian.json
var russian []byte

type localization struct {
	StartMessage       string `json:"start_message"`
	WaitMessage        string `json:"wait_message"`
	WaitScreenKeyboard struct {
		Find          string `json:"find"`
		ShowProfile   string `json:"show profile"`
		ChangeProfile string `json:"change profile"`
	} `json:"wait_screen_keyboard"`
	LikeScreenMessage  string `json:"like_screen_message"`
	LikeScreenKeyboard struct {
		Like    string `json:"like"`
		Dislike string `json:"dislike"`
		Report  string `json:"report"`
		Wait    string `json:"wait"`
	} `json:"like_screen_keyboard"`
	NoIndexedLeft    string `json:"no_indexed_left"`
	ProfileQuestions struct {
		Sex          string `json:"sex"`
		Man          string `json:"man"`
		NotMan       string `json:"not_man"`
		IncorrectSex string `json:"incorrect_sex"`
		Age          string `json:"age"`
		IncorrectAge string `json:"incorrect_age"`
		ProfileText  string `json:"profile_text"`
		Photo        string `json:"photo"`
		NoPhoto      string `json:"no_photo"`
		Finish       string `json:"finish"`
	} `json:"profile_questions"`
	Match struct {
		LikeScreen  string `json:"like_screen"`
		WaitScreen  string `json:"wait_screen"`
		Messeage    string `json:"messeage"`
		Next        string `json:"next"`
		FinishMatch string `json:"finish_match"`
	} `json:"match"`
}

var Russian localization

func init() {
	if err := json.Unmarshal(russian, &Russian); err != nil {
		log.Println(err)
	}
}
