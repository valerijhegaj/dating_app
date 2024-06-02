package state

import (
	"net/http"

	"date-app/internal/profile"
)

var State = make(map[int64]int)
var Data = make(map[int64]profile.Profile)
var UserID = make(map[int64]http.Client)
var IndexedID = make(map[int64]int)

const (
	State1 = iota
	State2
	State3
	State4
	State5
	State6
	State7
	State8
)
