package actual

import (
	"encoding/json"
	"io"
	"net/http"

	"date-app/internal/storage"
)

func GetHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		likes, err := db.GetMatches(r.Context(), userID, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(likes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

type requiredBody struct {
	UserID int `json:"user_id"`
}

func PostHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body requiredBody
		err = json.Unmarshal(data, &body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = db.DeleteNewMatch(r.Context(), userID, body.UserID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
