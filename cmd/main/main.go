package main

import (
	"flag"
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
	"date-app/internal/storage/postgres"
)

const (
	port = 8080
)

func handler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

func main() {
	var PORT int
	flag.IntVar(
		&PORT, "port", port, "port for server",
	)
	flag.Parse()

	db, err := postgres.New()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /endpoint/{id}", handler)
	http.HandleFunc("POST /endpoint", handler)
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
		"GET /api/v1/indexed", auth.CheckAuth(db)(indexed.Handler(db)),
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
		"DELETE /api/v1/matches/actual",
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
