package storage

import (
	"context"
	"errors"

	"date-app/internal/profile"
)

var (
	existUser               = errors.New("user already exist")
	ErrTokenNotFound        = errors.New("token not found")
	ErrLoginOrPasswordWrong = errors.New("login or password wrong")
	ErrLikeNotIndexed       = errors.New("like not indexed")
)

type Storage interface {
	Ping() error
	CreateUser(
		ctx context.Context, login, password, phoneNumber, email string,
	) error

	// CheckToken return nil if userID has tokenHash
	// return ErrTokenNotFound if token not found
	CheckToken(ctx context.Context, userID int, tokenHash string) error

	// CheckPassword return userID if correct login and passwordHash
	// returns ErrLoginOrPasswordWrong if incorrect login or passwordHash
	CheckPassword(
		ctx context.Context, login, passwordHash string,
	) (int, error)

	// AddToken return nil if everything ok and error if something went wrong
	AddToken(
		ctx context.Context, userID int, token string, maxAge int,
	) error

	// GetProfile - get profile by userID
	GetProfile(ctx context.Context, userID int) (profile.Profile, error)
	// AddProfile - add profile to userID
	AddProfile(ctx context.Context, userID int, p profile.Profile) error

	GetIndexed(ctx context.Context, userID int) (int, error)

	// GetLikes if isUserLikes = true return likes from userID
	// else return likes to userID
	GetLikes(
		ctx context.Context, userID int, isUsersLikes bool,
	) ([]profile.Like, error)

	GetMatches(ctx context.Context, userID int, viewed bool) (
		[]profile.Like, error,
	)
	DeleteNewMatch(ctx context.Context, userID, likedUserID int) error

	MakeLike(
		ctx context.Context, userID, likedUserID int, isLike bool,
	) (bool, error)
}
