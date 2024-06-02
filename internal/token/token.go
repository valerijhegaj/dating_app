package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"

	"date-app/configs"
)

func Generate256() string {
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		log.Println(err)
	}
	return hex.EncodeToString(b)[:256]
}

func GetFromCookie(cookie *http.Cookie) (string, int, error) {
	c := cookie.Value
	if len(c) < 256 {
		return "", 0, errors.New("bad cookie")
	}
	token, userID := c[:256], c[256:]
	ID, err := strconv.Atoi(userID)
	return token, ID, err
}

func ToCookie(token string, userID int) *http.Cookie {
	return &http.Cookie{
		Name: "token", Value: token + strconv.Itoa(userID),
		MaxAge: configs.Config.Main.MaxAge,
	}
}
