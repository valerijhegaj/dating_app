package bot_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"date-app/configs"
	"date-app/internal/profile"
	"date-app/internal/token"
)

var URL = "http://" + configs.Config.TgBot.Host + ":" + strconv.Itoa(configs.Config.Main.Port)

func CreateUser(login, password string) (http.Client, int, error) {
	const op = "CreateUser"

	client := http.Client{}
	client.Jar, _ = cookiejar.New(nil)
	body := fmt.Sprintf(
		`{"login":"%s", "password":"%s"}`, login, password,
	)

	r, err := client.Post(
		URL+"/api/v1/user", "application/json",
		bytes.NewBufferString(body),
	)
	if err != nil {
		return client, 0, fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return client, 0, fmt.Errorf("%s: %w", op, err)
	}

	r, err = client.Post(
		URL+"/api/v1/session", "application/json",
		bytes.NewBufferString(body),
	)
	if err != nil {
		return client, 0, fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return client, 0, fmt.Errorf("%s: %w", op, err)
	}
	u, err := url.Parse(URL)
	if err != nil {
		return client, 0, fmt.Errorf("%s: %w", op, err)
	}
	client.Jar.SetCookies(u, r.Cookies())
	_, ID, err := token.GetFromCookie(r.Cookies()[0])
	if err != nil {
		return client, ID, fmt.Errorf("%s: %w", op, err)
	}
	return client, ID, nil
}

func getClientID(client http.Client) (int, error) {
	const op = "getClientID"

	u, err := url.Parse(URL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if client.Jar == nil {
		return 0, fmt.Errorf("%s: bad client", op)
	}
	_, ID, err := token.GetFromCookie(client.Jar.Cookies(u)[0])
	if err != nil {
		return ID, fmt.Errorf("%s: %w", op, err)
	}
	return ID, nil
}

func UpdateProfile(
	client http.Client, userProfile profile.Profile,
) error {
	const op = "SendProfile"

	ID, err := getClientID(client)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	profileData, err := json.Marshal(userProfile)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	r, err := client.Post(
		URL+"/api/v1/profile/"+strconv.Itoa(ID), "",
		bytes.NewReader(profileData),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func PostLike(
	client http.Client, likeID int, isLike bool,
) (profile.Like, error) {
	const op = "PostLike"
	var like profile.Like

	endpoint := URL + "/api/v1/like/" + strconv.Itoa(likeID) + "?is_like="
	if isLike {
		endpoint += "1"
	} else {
		endpoint += "0"
	}

	r, err := client.Post(
		endpoint, "application/json",
		bytes.NewReader(nil),
	)
	if err != nil {
		return like, fmt.Errorf("%s: %w", op, err)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		_ = r.Body.Close()
		return like, fmt.Errorf("%s: %w", op, err)
	}
	err = r.Body.Close()
	if err != nil {
		return like, fmt.Errorf("%s: %w", op, err)
	}
	err = json.Unmarshal(body, &like)
	if err != nil {
		return like, fmt.Errorf("%s: %w", op, err)
	}
	return like, nil
}

func GetLikes(client http.Client) ([]profile.Like, error) {
	const op = "GetLikes"
	var likes []profile.Like

	r, err := client.Get(URL + "/api/v1/matches/actual")
	if err != nil {
		return likes, fmt.Errorf("%s: %w", op, err)
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return likes, fmt.Errorf("%s: %w", op, err)
	}
	if err = json.Unmarshal(body, &likes); err != nil {
		return likes, fmt.Errorf("%s: %w", op, err)
	}
	return likes, nil
}

func GetProfile(client http.Client, userID int) (
	profile.Profile, error,
) {
	const op = "GetProfile"
	var p profile.Profile

	r, err := client.Get(URL + "/api/v1/profile/" + strconv.Itoa(userID))
	if err != nil {
		return p, fmt.Errorf("%s: %w", op, err)
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return p, fmt.Errorf("%s: %w", op, err)
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		return p, fmt.Errorf("%s: %w", op, err)
	}
	return p, nil
}

func PostProfileViewed(client http.Client, userID int) error {
	const op = "PostProfileViewed"

	data, err := json.Marshal(profile.Like{UserID: userID})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	r, err := client.Post(
		URL+"/api/v1/matches/actual", "",
		bytes.NewReader(data),
	)
	defer r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func GetIndexed(client http.Client) (int, error) {
	const op = "GetIndexed"

	r, err := client.Get(URL + "/api/v1/indexed")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if r.StatusCode == http.StatusForbidden {
		return 0, nil
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var x struct {
		UserID int `json:"user_id"`
	}
	err = json.Unmarshal(body, &x)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return x.UserID, nil
}
