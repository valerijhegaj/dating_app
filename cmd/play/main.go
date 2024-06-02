package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"date-app/configs"
)

func main() {
	r, err := http.Post(
		"http://"+
			configs.Config.TgBot.Host+":"+
			strconv.Itoa(configs.Config.Main.Port)+"/api/v1/session",
		"", bytes.NewReader([]byte(`{"login":"test7"}`)),
	)
	fmt.Println(r, err)
	c := http.Client{}
	r, err = c.Post(
		"http://"+
			configs.Config.TgBot.Host+":"+
			strconv.Itoa(configs.Config.Main.Port)+"/api/v1/session",
		"", bytes.NewReader([]byte(`{"login":"test7"}`)),
	)
	fmt.Println(
		r, err,
		r.Cookies(),
	)
	u, err := url.Parse(
		"http://" +
			configs.Config.TgBot.Host + ":" +
			strconv.Itoa(configs.Config.Main.Port),
	)
	fmt.Println(u, err)
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(u, r.Cookies())
	c.Jar = jar
	fmt.Println(r, err)
	r, err = c.Get(
		"http://" +
			configs.Config.TgBot.Host + ":" +
			strconv.Itoa(configs.Config.Main.Port) + "/api/v1/likes/my",
	)
	fmt.Println(r, err)
}
