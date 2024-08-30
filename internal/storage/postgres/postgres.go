package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

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
	const op = "storage.postgres.MakeLike"
	query := `SELECT dating_data.make_like($1, $2, $3)`
	var isAllowed bool
	err := s.Db.QueryRowContext(
		ctx, query, userID, likedUserID, isLike,
	).Scan(&isAllowed)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if !isAllowed {
		return false, fmt.Errorf("%s: %w", op, storage.ErrLikeNotIndexed)
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
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (s *Storage) GetMatches(
	ctx context.Context, userID int, viewed bool,
) ([]profile.Like, error) {
	const op = "storage.postgres.GetMatches"
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
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var matches []profile.Like
	for rows.Next() {
		var like profile.Like
		if err = rows.Scan(&like.UserID, &like.Time); err != nil {
			return matches, fmt.Errorf("%s: %w", op, err)
		}
		matches = append(matches, like)
	}
	if err = rows.Err(); err != nil {
		return matches, err
	}
	if err != nil {
		return matches, fmt.Errorf("%s: %w", op, err)
	}
	return matches, nil
}

func (s *Storage) DeleteNewMatch(
	ctx context.Context, userID, likedUserID int,
) error {
	const op = "storage.postgres.DeleteNewMatch"
	query := `UPDATE dating_data.starred_users 
            SET viewed = true 
            WHERE user_id = $1 AND starred_user_id = $2`
	_, err := s.Db.ExecContext(ctx, query, userID, likedUserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetLikes(
	ctx context.Context, userID int, isUsersLikes bool,
) ([]profile.Like, error) {
	const op = "storage.postgres.GetLikes"
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
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer r.Close()
	var likes []profile.Like
	for r.Next() {
		var like profile.Like
		if err = r.Scan(&like.UserID, &like.Time); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		likes = append(likes, like)
	}
	if err = r.Err(); err != nil {
		return likes, fmt.Errorf("%s: %w", op, err)
	}
	return likes, nil
}

func (s *Storage) GetIndexed(ctx context.Context, userID int) (
	int, error,
) {
	const op = "storage.postgres.GetIndexed"
	var indexedID int
	if err := s.Db.QueryRowContext(
		ctx,
		`SELECT indexed_user_id FROM dating_data.indexed_users WHERE user_id = $1 LIMIT 1`,
		userID,
	).Scan(&indexedID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return indexedID, nil
}

func (s *Storage) GetProfile(
	ctx context.Context, userID int,
) (profile.Profile, error) {
	const op = "storage.postgres.GetProfile"
	var userProfile profile.Profile
	var profileID int
	query := `SELECT profile_id, profile_text, sex, birthday, name, url
				   FROM dating_data.profile
				   WHERE user_id = $1`
	err := s.Db.QueryRowContext(ctx, query, userID).Scan(
		&profileID, &userProfile.ProfileText, &userProfile.Sex,
		&userProfile.Birthday, &userProfile.Name,
		&userProfile.URL,
	)
	if err != nil {
		return userProfile, fmt.Errorf("%s: %w", op, err)
	}
	query = `SELECT image_url
					 FROM dating_data.profile_photo
					 LEFT JOIN dating_data.photo
				   ON profile_photo.photo_id = photo.photo_id
				   WHERE profile_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, profileID)
	if err != nil {
		return userProfile, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var URL string
		err = rows.Scan(&URL)
		if err != nil {
			return userProfile, fmt.Errorf("%s: %w", op, err)
		}
		userProfile.Photo = append(userProfile.Photo, URL)
	}
	if err = rows.Err(); err != nil {
		return userProfile, fmt.Errorf("%s: %w", op, err)
	}
	return userProfile, nil
}

func (s *Storage) AddProfile(
	ctx context.Context, userID int, profile profile.Profile,
) error {
	const op = "storage.postgres.AddProfile"
	var profileID int
	if err := s.Db.QueryRowContext(
		ctx, `SELECT dating_data.create_profile($1, $2, $3, $4, $5, $6)`,
		userID, profile.ProfileText, profile.Sex, profile.Birthday,
		profile.Name, profile.URL,
	).Scan(&profileID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for _, url := range profile.Photo {
		_, err := s.Db.ExecContext(
			ctx, `CALL dating_data.add_new_photo($1, $2)`, profileID, url,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func (s *Storage) AddToken(
	ctx context.Context, userID int, token string, maxAge int,
) error {
	const op = "storage.postgres.AddToken"
	query := `INSERT INTO dating_data.auth
						(user_id, token_hash, end_time) VALUES 
						($1, $2, NOW() + $3 * interval '1 second')`
	_, err := s.Db.ExecContext(
		ctx, query,
		userID, token, maxAge,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) CheckPassword(
	ctx context.Context, login string, passwordHash string,
) (int, error) {
	const op = "storage.postgres.CheckPassword"
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
		return 0, fmt.Errorf(
			"%s: %w", op, storage.ErrLoginOrPasswordWrong,
		)
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if passwordHash != realPasswordHash {
		return 0, fmt.Errorf(
			"%s: %w", op, storage.ErrLoginOrPasswordWrong,
		)
	}
	return userID, nil
}

func (s *Storage) CheckToken(
	ctx context.Context, userID int, tokenHash string,
) error {
	const op = "storage.postgres.CheckToken"
	query := `SELECT token_hash 
						FROM dating_data.auth 
						WHERE user_id = $1 AND token_hash = $2`
	err := s.Db.QueryRowContext(
		ctx, query,
		userID, tokenHash,
	).Scan(&tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, storage.ErrTokenNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) CreateUser(
	ctx context.Context, login, password, phoneNumber, email string,
) error {
	const op = "storage.postgres.CreateUser"
	query := `INSERT INTO dating_data.user 
						(login, password_hash, phone_number, email) VALUES 
						($1, $2, $3, $4)`
	_, err := s.Db.ExecContext(
		ctx, query,
		login, password, phoneNumber, email,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetAllUserIDs(ctx context.Context) ([]int, error) {
	const op = "storage.postgres.GetAllUserIDs"
	var userIDs []int

	getUserIDs := `SELECT user_id
								 FROM dating_data.profile;`
	rows, err := s.Db.QueryContext(ctx, getUserIDs)
	if err != nil {
		return userIDs, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		err = rows.Scan(&userID)
		if err != nil {
			return userIDs, fmt.Errorf("%s: %w", op, err)
		}
		userIDs = append(userIDs, userID)
	}
	if err = rows.Err(); err != nil {
		return userIDs, fmt.Errorf("%s: %w", op, err)
	}
	return userIDs, nil
}

func (s *Storage) GetNewLikes(ctx context.Context, userID int) (
	[]int, error,
) {
	const op = "storage.postgres.GetNewLikes"
	var likedUserIDs []int

	getLiked := `SELECT user_id
						   FROM dating_data.starred_users
							 WHERE starred_user_id = $1
                 AND is_liked = true
                 AND user_id NOT IN (
    							SELECT starred_user_id
    							FROM dating_data.starred_users
    							WHERE user_id = $1);`
	rows, err := s.Db.QueryContext(ctx, getLiked, userID)
	if err != nil {
		return likedUserIDs, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pairID int
		err = rows.Scan(&pairID)
		if err != nil {
			return likedUserIDs, fmt.Errorf("%s: %w", op, err)
		}
		likedUserIDs = append(likedUserIDs, pairID)
	}
	if err = rows.Err(); err != nil {
		return likedUserIDs, fmt.Errorf("%s: %w", op, err)
	}
	return likedUserIDs, nil
}

func (s *Storage) GetNewPairs(
	ctx context.Context, userID int, n int,
) ([]int, error) {
	const op = "storage.postgres.GetNewPairs"
	var newPairIDs []int

	userProfile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return newPairIDs, fmt.Errorf("%s: %w", op, err)
	}
	pairSex := !userProfile.Sex

	getNewPairs := `SELECT profile.user_id
									FROM dating_data.profile
									LEFT JOIN dating_data.user 
									ON profile.user_id = "user".user_id
									WHERE profile.user_id != $1
  									AND sex = $3
  									AND profile.user_id NOT IN (
    									SELECT starred_user_id
    									FROM dating_data.starred_users
    									WHERE user_id = $1
  									)
  									AND profile.user_id NOT IN (
      								SELECT user_id
      								FROM dating_data.starred_users
      								WHERE starred_user_id = $1
  									)
									ORDER BY last_online DESC
									LIMIT $2;`
	rows, err := s.Db.QueryContext(ctx, getNewPairs, userID, n, pairSex)
	if err != nil {
		return newPairIDs, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pairID int
		err = rows.Scan(&pairID)
		if err != nil {
			return newPairIDs, fmt.Errorf("%s: %w", op, err)
		}
		newPairIDs = append(newPairIDs, pairID)
	}
	if err = rows.Err(); err != nil {
		return newPairIDs, fmt.Errorf("%s: %w", op, err)
	}
	return newPairIDs, nil
}

func (s *Storage) LoadIndexed(
	ctx context.Context, userID int, indexedIDs []int,
) error {
	const op = "storage.postgres.GetNewPairs"

	if len(indexedIDs) == 0 {
		return nil
	}

	removeIndexed := `DELETE FROM dating_data.indexed_users 
										WHERE user_id = $1;`
	_, err := s.Db.ExecContext(ctx, removeIndexed, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	loadIndexed := `INSERT INTO dating_data.indexed_users (user_id, indexed_user_id) VALUES `
	for i := range indexedIDs {
		if i == 0 {
			loadIndexed += "($1, $2)"
			continue
		}
		loadIndexed += ", ($1, $" + strconv.Itoa(i+2) + ")"
	}
	loadIndexed += ";"

	args := append([]any{}, userID)
	for _, indexedID := range indexedIDs {
		args = append(args, indexedID)
	}

	if _, err = s.Db.ExecContext(
		ctx, loadIndexed, args...,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
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
