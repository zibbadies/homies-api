package models

import "time"

// import "time"
var NoTime = time.Time{}

type List struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Item struct {
	Id        string  `json:"id"`
	Text      string  `json:"text"`
	Completed bool    `json:"completed"`
	Author    string  `json:"author"`
	CreatedAt string  `json:"created_at"`
	DueAt     *string `json:"due_at"`
}

// One calendar day (UTC) with due tasks, for the due-days preview endpoint
type DueDay struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
	Open  int    `json:"open"`
}
