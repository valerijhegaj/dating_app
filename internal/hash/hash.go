package hash

import (
	"crypto/sha256"
	"fmt"

	"date-app/configs"
)

func Calculate(data string) string {
	h := sha256.New()
	h.Write([]byte(data))

	return fmt.Sprintf(
		"%x", h.Sum([]byte(configs.Config.Main.HashKey)),
	)
}
