package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"date-app/internal/indexer"
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

func PostHandler(
	db storage.Storage, Indexer indexer.Indexer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")
		sessionUserID := r.Context().Value("userID").(int)
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
		ProfileToChangeUserID, err := strconv.Atoi(userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// forbidden when other user change profile
		if ProfileToChangeUserID != sessionUserID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err = db.AddProfile(
			r.Context(), sessionUserID, p,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err = Indexer.IndexUser(
			r.Context(), sessionUserID,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
