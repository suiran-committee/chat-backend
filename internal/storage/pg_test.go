package storage

import (
	"context"
	"regexp"
	"testing"

	"chat-backend/internal/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestPGStore_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	st := &pgStore{db: db}
	msg := model.Message{Username: "bob", Content: "hello"}

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO messages (username, content) VALUES ($1, $2)`)).
		WithArgs(msg.Username, msg.Content).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := st.Save(context.Background(), msg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPGStore_Last(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	st := &pgStore{db: db}

	rows := sqlmock.NewRows([]string{"username", "content"}).
		AddRow("bob", "yo").
		AddRow("alice", "hi")
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT username, content FROM messages ORDER BY id DESC LIMIT $1`)).
		WithArgs(2).WillReturnRows(rows)

	msgs, err := st.Last(context.Background(), 2)
	if err != nil {
		t.Fatalf("Last: %v", err)
	}

	if len(msgs) != 2 ||
		msgs[0].Username != "alice" || msgs[0].Content != "hi" ||
		msgs[1].Username != "bob" || msgs[1].Content != "yo" {
		t.Fatalf("unexpected result: %+v", msgs)
	}
}
