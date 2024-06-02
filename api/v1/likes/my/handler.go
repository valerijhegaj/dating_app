package my

import (
	"encoding/json"
	"net/http"

	"date-app/internal/storage"
)

func Handler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		likes, err := db.GetLikes(r.Context(), userID, true)
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
