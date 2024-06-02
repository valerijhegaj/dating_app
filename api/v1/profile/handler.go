package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"date-app/internal/profile"
	"date-app/internal/storage"
)

func GetHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.PathValue("user_id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		profile, err := db.GetProfile(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, err := json.Marshal(profile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

func PostHandler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")
		ID := r.Context().Value("userID").(int)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var p profile.Profile
		if err = json.Unmarshal(body, &p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		uID, err := strconv.Atoi(userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if uID != ID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err = db.AddProfile(r.Context(), ID, p); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
