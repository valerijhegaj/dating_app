package main

import (
	"fmt"
	"log"
	"net/http"

	"date-app/api/middleware/auth"
	"date-app/api/v1/indexed"
	"date-app/api/v1/like"
	"date-app/api/v1/likes/me"
	"date-app/api/v1/likes/my"
	"date-app/api/v1/matches"
	"date-app/api/v1/matches/actual"
	"date-app/api/v1/profile"
	"date-app/api/v1/session"
	"date-app/api/v1/user"
	"date-app/configs"
	"date-app/internal/storage/postgres"
)

func main() {
	PORT := configs.Config.Main.Port

	db, err := postgres.New()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("POST /api/v1/user", user.Handler(db))

	http.Handle(
		"GET /api/v1/session", auth.CheckAuth(db)(session.GetHandler),
	)
	http.Handle("POST /api/v1/session", session.PostHandler(db))

	http.Handle("GET /api/v1/profile/{user_id}", profile.GetHandler(db))
	http.Handle(
		"POST /api/v1/profile/{user_id}",
		auth.CheckAuth(db)(profile.PostHandler(db)),
	)

	http.Handle(
		"GET /api/v1/indexed", auth.CheckAuth(db)(indexed.GetHandler(db)),
	)

	http.Handle(
		"GET /api/v1/likes/my", auth.CheckAuth(db)(my.Handler(db)),
	)
	http.Handle(
		"GET /api/v1/likes/me", auth.CheckAuth(db)(me.Handler(db)),
	)

	http.Handle(
		"GET /api/v1/matches", auth.CheckAuth(db)(matches.Handler(db)),
	)

	http.Handle(
		"GET /api/v1/matches/actual",
		auth.CheckAuth(db)(actual.GetHandler(db)),
	)
	http.Handle(
		"POST /api/v1/matches/actual",
		auth.CheckAuth(db)(actual.DeleteHandler(db)),
	)

	http.Handle(
		"POST /api/v1/like/{user_id}",
		auth.CheckAuth(db)(like.Handler(db)),
	)

	if err = http.ListenAndServe(
		fmt.Sprintf(":%d", PORT), nil,
	); err != nil {
		log.Fatal(err.Error())
	}
}
