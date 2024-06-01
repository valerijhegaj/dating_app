package auth

import (
	"context"
	"errors"
	"net/http"

	"date-app/internal/hash"
	"date-app/internal/storage"
	"date-app/internal/token"
)

type TokenChecker interface {
	CheckToken(ctx context.Context, userID int, tokenHash string) error
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return (http.HandlerFunc)(
			func(w http.ResponseWriter, r *http.Request) {
				cookie, err := r.Cookie("token")
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				tok, userID, err := token.GetFromCookie(cookie)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				tokenHash := hash.Calculate(tok)
				err = checker.CheckToken(r.Context(), userID, tokenHash)
				if errors.Is(err, storage.ErrTokenNotFound) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				newCtx := context.WithValue(r.Context(), "userID", userID)
				requestWithUser := r.WithContext(newCtx)
				next.ServeHTTP(w, requestWithUser)
			},
		)
	}
}
