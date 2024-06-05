package like

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"date-app/internal/profile"
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
		isLike, err := strconv.Atoi(r.URL.Query().Get("is_like"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		isMatch, err := db.MakeLike(
			r.Context(), userID, likeUserID, isLike == 1,
		)
		if err != nil {
			if errors.Is(err, storage.ErrLikeNotIndexed) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		var ans profile.Like
		if isMatch {
			ans.UserID = likeUserID
			ans.Time = time.Now().Format("2006-01-02")
		}
		data, _ := json.Marshal(ans)
		_, _ = w.Write(data)
	}
}
