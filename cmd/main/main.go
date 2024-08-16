package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if err = db.Ping(); err != nil {
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
		auth.CheckAuth(db)(actual.PostHandler(db)),
	)

	http.Handle(
		"POST /api/v1/like/{user_id}",
		auth.CheckAuth(db)(like.Handler(db)),
	)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", PORT),
	}

	shutdownChan := make(chan struct{})

	go func() {
		if err = server.ListenAndServe(); !errors.Is(
			err, http.ErrServerClosed,
		) {
			log.Fatal(err.Error())
		}
		log.Println("ListenAndServe stopped.")
		shutdownChan <- struct{}{}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer shutdownRelease()

	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	<-shutdownChan
	log.Println("Server stopped.")

}
