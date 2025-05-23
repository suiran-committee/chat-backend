package testutil

import (
	"context"

	"chat-backend/internal/model"
	"chat-backend/internal/storage"
)

type MockStore struct {
	SaveFn func(ctx context.Context, m model.Message) error
	LastFn func(ctx context.Context, n int) ([]model.Message, error)
}

var _ storage.HistoryStore = (*MockStore)(nil)

func (m *MockStore) Save(ctx context.Context, msg model.Message) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, msg)
	}
	return nil
}
func (m *MockStore) Last(ctx context.Context, n int) ([]model.Message, error) {
	if m.LastFn != nil {
		return m.LastFn(ctx, n)
	}
	return nil, nil
}
func (m *MockStore) Close() error { return nil }
