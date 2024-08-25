package handler_bot

import (
	"net/http"

	"date-app/internal/profile"
)

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

var Manager = MemoryManager{
	userState:   make(map[int64]int),
	userProfile: make(map[int64]profile.Profile),
	userClient:  make(map[int64]http.Client),
	indexedID:   make(map[int64]int),
	tgUserID:    make(map[int]int64),
}

type MemoryManager struct {
	userState   map[int64]int
	userProfile map[int64]profile.Profile
	userClient  map[int64]http.Client
	indexedID   map[int64]int
	tgUserID    map[int]int64
}

func (m *MemoryManager) UpdateState(tgUserID int64, state int) {
	m.userState[tgUserID] = state
}

func (m *MemoryManager) GetState(tgUserID int64) int {
	st, ok := m.userState[tgUserID]
	if !ok {
		return StateNonAuthed
	}
	return st
}

func (m *MemoryManager) UpdateProfile(
	tgUserID int64, p profile.Profile,
) {
	m.userProfile[tgUserID] = p
}

func (m *MemoryManager) UpdateProfileText(
	tgUserID int64, profileText string,
) {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		m.userProfile[tgUserID] = profile.Profile{ProfileText: profileText}
	} else {
		p.ProfileText = profileText
		m.userProfile[tgUserID] = p
	}
}

func (m *MemoryManager) UpdateProfileSex(tgUserID int64, sex bool) {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		m.userProfile[tgUserID] = profile.Profile{Sex: sex}
	} else {
		p.Sex = sex
		m.userProfile[tgUserID] = p
	}
}

func (m *MemoryManager) UpdateProfileBirthday(
	tgUserID int64, birthday string,
) {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		m.userProfile[tgUserID] = profile.Profile{Birthday: birthday}
	} else {
		p.Birthday = birthday
		m.userProfile[tgUserID] = p
	}
}

func (m *MemoryManager) UpdateProfileName(
	tgUserID int64, name string,
) {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		m.userProfile[tgUserID] = profile.Profile{Name: name}
	} else {
		p.Name = name
		m.userProfile[tgUserID] = p
	}
}

func (m *MemoryManager) UpdateProfilePhoto(
	tgUserID int64, photo string,
) {
	userProfile, ok := m.userProfile[tgUserID]
	if !ok {
		var p profile.Profile
		p.Photo = append(p.Photo, photo)
		m.userProfile[tgUserID] = p
		return
	}
	userProfile.Photo = append(userProfile.Photo, photo)
	m.userProfile[tgUserID] = userProfile
}

func (m *MemoryManager) UpdateProfileURL(tgUserID int64, URL string) {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		m.userProfile[tgUserID] = profile.Profile{URL: URL}
	} else {
		p.URL = URL
		m.userProfile[tgUserID] = p
	}
}

func (m *MemoryManager) GetProfile(tgUserID int64) profile.Profile {
	p, ok := m.userProfile[tgUserID]
	if !ok {
		return profile.Profile{}
	}
	delete(m.userProfile, tgUserID)
	return p
}

func (m *MemoryManager) UpdateClient(
	tgUserID int64, client http.Client,
) {
	m.userClient[tgUserID] = client
}

func (m *MemoryManager) CheckClient(tgUserID int64) bool {
	_, ok := m.userClient[tgUserID]
	return ok
}

func (m *MemoryManager) GetClient(tgUserID int64) http.Client {
	client, ok := m.userClient[tgUserID]
	if !ok {
		return http.Client{}
	}
	return client
}

func (m *MemoryManager) UpdateIndexed(tgUserID int64, indexedID int) {
	m.indexedID[tgUserID] = indexedID
}

func (m *MemoryManager) GetIndexed(tgUserID int64) int {
	indexedID, ok := m.indexedID[tgUserID]
	if !ok {
		return 0
	}
	return indexedID
}

func (m *MemoryManager) UpdateTgUserID(tgUserID int64, userID int) {
	m.tgUserID[userID] = tgUserID
}

func (m *MemoryManager) GetTgUserID(userID int) int64 {
	tgUserID, ok := m.tgUserID[userID]
	if !ok {
		return 0
	}
	return tgUserID
}
