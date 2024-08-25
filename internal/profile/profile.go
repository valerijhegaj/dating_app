package profile

type Profile struct {
	ProfileText string   `json:"profile_text"`
	Sex         bool     `json:"sex"`
	Birthday    string   `json:"birthday"`
	Name        string   `json:"name"`
	Photo       []string `json:"photo"`
	URL         string   `json:"url"`
}

type Like struct {
	UserID int    `json:"user_id"`
	Time   string `json:"time"`
}
