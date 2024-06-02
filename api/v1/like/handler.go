package like

import (
	"net/http"
	"strconv"

	"date-app/internal/storage"
)

func Handler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		likeUserID, err := strconv.Atoi(r.PathValue("user_id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = db.MakeLike(r.Context(), userID, likeUserID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
