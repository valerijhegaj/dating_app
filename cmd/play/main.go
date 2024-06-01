package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		log.Println(err)
	}
	fmt.Println(hex.EncodeToString(b))

}
