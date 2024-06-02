package session

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"date-app/configs"
	"date-app/internal/hash"
	"date-app/internal/storage"
	"date-app/internal/token"
)

var GetHandler http.HandlerFunc = func(
	w http.ResponseWriter, r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}

type requiredBody struct {
	Login    string
	Password string
}

func PostHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body requiredBody
		if err = json.Unmarshal(data, &body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userID, err := db.CheckPassword(
			r.Context(), body.Login, hash.Calculate(body.Password),
		)
		if errors.Is(err, storage.ErrLoginOrPasswordWrong) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tok := token.Generate256()
		err = db.AddToken(
			r.Context(), userID, hash.Calculate(tok),
			configs.Config.Main.MaxAge,
		)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cookie := token.ToCookie(tok, userID)
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusCreated)
	}
}
