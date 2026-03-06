package data

import (
	"context"
	"database/sql"
	"time"
)

type BookStore struct {
	DB *sql.DB
}

func (s *BookStore) GetAll() ([]Book, error) {
	const query = `SELECT id, title, author, year FROM books ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book

	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
