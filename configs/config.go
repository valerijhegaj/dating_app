package configs

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed config.json
var config []byte

type main struct {
	Port    int    `json:"port"`
	HashKey string `json:"hash_key"`
	MaxAge  int    `json:"max_age"`
}

type database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type tgBot struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

type ConfigType struct {
	Main     main     `json:"main"`
	Database database `json:"database"`
	TgBot    tgBot    `json:"tg-bot"`
}

var Config ConfigType

func init() {
	if err := json.Unmarshal(
		config, &Config,
	); err != nil {
		fmt.Println(err)
	}
}
