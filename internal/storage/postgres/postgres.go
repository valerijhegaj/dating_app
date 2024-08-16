package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"date-app/configs"
	"date-app/internal/profile"
	"date-app/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	Db *sql.DB
}

func (s *Storage) MakeLike(
	ctx context.Context, userID, likedUserID int, isLike bool,
) (bool, error) {
	query := `SELECT dating_data.make_like($1, $2, $3)`
	var isAllowed bool
	err := s.Db.QueryRowContext(
		ctx, query, userID, likedUserID, isLike,
	).Scan(&isAllowed)
	if err != nil {
		return false, err
	}
	if !isAllowed {
		return false, storage.ErrLikeNotIndexed
	}
	if !isLike {
		return false, nil
	}
	query = `SELECT user_id 
           FROM dating_data.starred_users 
					 WHERE user_id=$1 AND starred_user_id=$2 AND is_liked = true`
	err = s.Db.QueryRowContext(
		ctx, query, likedUserID, userID,
	).Scan(&likedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Storage) GetMatches(
	ctx context.Context, userID int, viewed bool,
) ([]profile.Like, error) {
	query := `SELECT l.starred_user_id, 
            CASE WHEN l.time > r.time 
                THEN l.time 
                ELSE r.time 
            END
						FROM (SELECT * 
						      FROM dating_data.starred_users 
						      WHERE user_id = $1 AND is_liked = true) as l
						JOIN (SELECT user_id, starred_user_id, time 
						      FROM dating_data.starred_users 
						      WHERE is_liked = true) as r
						ON l.user_id = r.starred_user_id AND r.user_id = l.starred_user_id 
						WHERE l.viewed = $2`
	rows, err := s.Db.QueryContext(ctx, query, userID, viewed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ans []profile.Like
	for rows.Next() {
		var like profile.Like
		if err = rows.Scan(&like.UserID, &like.Time); err != nil {
			return ans, err
		}
		ans = append(ans, like)
	}
	if err = rows.Err(); err != nil {
		return ans, err
	}
	return ans, err
}

func (s *Storage) DeleteNewMatch(
	ctx context.Context, userID, likedUserID int,
) error {
	query := `UPDATE dating_data.starred_users 
            SET viewed = true 
            WHERE user_id = $1 AND starred_user_id = $2`
	_, err := s.Db.ExecContext(ctx, query, userID, likedUserID)
	return err
}

func (s *Storage) GetLikes(
	ctx context.Context, userID int, isUsersLikes bool,
) ([]profile.Like, error) {
	var query string
	if isUsersLikes {
		query = `SELECT starred_user_id, time 
             FROM dating_data.starred_users 
             WHERE user_id = $1 AND is_liked = true`
	} else {
		query = `SELECT user_id, time 
						 FROM dating_data.starred_users 
						 WHERE starred_user_id = $1 AND is_liked = true`
	}
	r, err := s.Db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var ans []profile.Like
	for r.Next() {
		var like profile.Like
		if err = r.Scan(&like.UserID, &like.Time); err != nil {
			return nil, err
		}
		ans = append(ans, like)
	}
	if err = r.Err(); err != nil {
		return ans, err
	}
	return ans, nil
}

func (s *Storage) GetIndexed(ctx context.Context, userID int) (
	int, error,
) {
	var indexedID int
	if err := s.Db.QueryRowContext(
		ctx,
		`SELECT indexed_user_id FROM dating_data.indexed_users WHERE user_id = $1 LIMIT 1`,
		userID,
	).Scan(&indexedID); err != nil {
		return 0, err
	}
	return indexedID, nil
}

func (s *Storage) GetProfile(
	ctx context.Context, userID int,
) (profile.Profile, error) {
	var ans profile.Profile
	var profileID int
	query := `SELECT profile_id, profile_text, sex, birthday, name, url
				   FROM dating_data.profile
				   WHERE user_id = $1`
	err := s.Db.QueryRowContext(ctx, query, userID).Scan(
		&profileID, &ans.ProfileText, &ans.Sex, &ans.Birthday, &ans.Name,
		&ans.URL,
	)
	if err != nil {
		return ans, err
	}
	query = `SELECT image_url
					 FROM dating_data.profile_photo
					 LEFT JOIN dating_data.photo
				   ON profile_photo.photo_id = photo.photo_id
				   WHERE profile_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, profileID)
	if err != nil {
		return ans, err
	}
	defer rows.Close()
	for rows.Next() {
		var URL string
		err = rows.Scan(&URL)
		if err != nil {
			return ans, err
		}
		ans.Photo = append(ans.Photo, URL)
	}
	if err = rows.Err(); err != nil {
		return ans, err
	}
	return ans, nil
}

func (s *Storage) AddProfile(
	ctx context.Context, userID int, profile profile.Profile,
) error {
	var profileID int
	if err := s.Db.QueryRowContext(
		ctx, `SELECT dating_data.create_profile($1, $2, $3, $4, $5, $6)`,
		userID, profile.ProfileText, profile.Sex, profile.Birthday,
		profile.Name, profile.URL,
	).Scan(&profileID); err != nil {
		return err
	}
	for _, url := range profile.Photo {
		_, err := s.Db.ExecContext(
			ctx, `CALL dating_data.add_new_photo($1, $2)`, profileID, url,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) AddToken(
	ctx context.Context, userID int, token string, maxAge int,
) error {
	query := `INSERT INTO dating_data.auth
						(user_id, token_hash, end_time) VALUES 
						($1, $2, NOW() + $3 * interval '1 second')`
	_, err := s.Db.ExecContext(
		ctx, query,
		userID, token, maxAge,
	)
	return err
}

func (s *Storage) CheckPassword(
	ctx context.Context, login string, passwordHash string,
) (int, error) {
	var realPasswordHash string
	var userID int
	query := `SELECT user_id, password_hash
						FROM dating_data.user
						WHERE login = $1`
	err := s.Db.QueryRowContext(
		ctx, query,
		login,
	).Scan(&userID, &realPasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrLoginOrPasswordWrong
	}
	if err != nil {
		return 0, err
	}
	if passwordHash != realPasswordHash {
		return 0, storage.ErrLoginOrPasswordWrong
	}
	return userID, nil
}

func (s *Storage) CheckToken(
	ctx context.Context, userID int, tokenHash string,
) error {
	query := `SELECT token_hash 
						FROM dating_data.auth 
						WHERE user_id = $1 AND token_hash = $2`
	err := s.Db.QueryRowContext(
		ctx, query,
		userID, tokenHash,
	).Scan(&tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrTokenNotFound
	}
	return err
}

func (s *Storage) CreateUser(
	ctx context.Context, login, password, phoneNumber, email string,
) error {
	query := `INSERT INTO dating_data.user 
						(login, password_hash, phone_number, email) VALUES 
						($1, $2, $3, $4)`
	_, err := s.Db.ExecContext(
		ctx, query,
		login, password, phoneNumber, email,
	)
	return err
}

func (s *Storage) Ping() error {
	return s.Db.Ping()
}

func (s *Storage) Close() error {
	return s.Db.Close()
}

var _ storage.Storage = (*Storage)(nil)

//go:embed init.sql
var initQuery string

//go:embed trigger.sql
var triggerQuery string

//go:embed functions.sql
var functionsQuery string

// TODO init config in secrets
var (
	host     = configs.Config.Database.Host
	port     = configs.Config.Database.Port
	user     = configs.Config.Database.User
	password = configs.Config.Database.Password
)

func New() (storage.Storage, error) {
	const op = "storage.postgres.New"
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s "+
			"password=%s sslmode=disable",
		host, port, user, password,
	)
	db, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(initQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(triggerQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(functionsQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db}, nil
}
