package storage

import (
	"context"
	"database/sql"
	"fmt"

	"chat-backend/internal/model"

	_ "github.com/lib/pq"
)

type HistoryStore interface {
	Save(ctx context.Context, m model.Message) error
	Last(ctx context.Context, n int) ([]model.Message, error)
	Close() error
}

type pgStore struct {
	db *sql.DB
}

func NewPostgres(host, user, pass, name string) (HistoryStore, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, pass, name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &pgStore{db: db}, nil
}

func (p *pgStore) Save(ctx context.Context, m model.Message) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO messages (username, content) VALUES ($1, $2)`,
		m.Username, m.Content)
	return err
}

func (p *pgStore) Last(ctx context.Context, n int) ([]model.Message, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT username, content FROM messages ORDER BY id DESC LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.Username, &m.Content); err != nil {
			return nil, err
		}
		out = append([]model.Message{m}, out...)
	}
	return out, rows.Err()
}

func (p *pgStore) Close() error { return p.db.Close() }
