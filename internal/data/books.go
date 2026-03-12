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

func (s *BookStore) Get(id int64) (*Book, error) {
	if id < 1 {
		return nil, sql.ErrNoRows
	}

	const query = `SELECT id, title, author, year FROM books WHERE id = ?`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var book Book

	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
	)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (s *BookStore) Insert(book *Book) (*Book, error) {
	query := `INSERT INTO books (title, author, year) VALUES (?, ?, ?)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := s.DB.ExecContext(ctx, query, book.Title, book.Author, book.Year)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	book.ID = id

	return book, nil
}
