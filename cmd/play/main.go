package main

import (
	"strconv"
)

func main() {
	userID := 1
	indexedIDs := []any{2, 3, 4, 5, 6, 7, 8, 9, 10}
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
	args = append(args, indexedIDs...)
	x := 1 + 1
	args[0] = x
}
