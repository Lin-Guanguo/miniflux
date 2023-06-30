package model // import "miniflux.app/model"

import "time" // Report  represents user report data
type Report struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ReportRequest struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
