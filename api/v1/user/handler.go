package user

import (
	"encoding/json"
	"io"
	"net/http"

	"date-app/internal/hash"
	"date-app/internal/storage"
)

type requiredBody struct {
	Login       string
	Password    string
	PhoneNumber string `json:"phone_number"`
	Email       string
}

//type storage interface {
//	CreateUser(
//		ctx context.Context, login string, password string,
//		phoneNumber string, email string,
//	) error
//}

func Handler(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var data requiredBody
		if err = json.Unmarshal(body, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		passwordHash := hash.Calculate(data.Password)
		if err = db.CreateUser(
			r.Context(), data.Login, passwordHash,
			data.PhoneNumber,
			data.Email,
		); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)

	}
}
