package model

// Note the note model
type Note struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Hidden    bool   `json:"hidden"`
	Encrypted bool   `json:"encrypted"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
