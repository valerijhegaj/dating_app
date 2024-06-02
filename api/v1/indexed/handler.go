package indexed

import (
	"encoding/json"
	"net/http"

	"date-app/internal/storage"
)

type ans struct {
	UserID int `json:"user_id"`
}

func GetHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		indexedID, err := db.GetIndexed(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		data, err := json.Marshal(ans{UserID: indexedID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
