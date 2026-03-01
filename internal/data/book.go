package data

type Book struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author,omitempty"`
	Year   int    `json:"year,omitempty"`
}
