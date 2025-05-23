package testutil

import (
	"context"

	"chat-backend/internal/model"
	"chat-backend/internal/pubsub"
)

type MockPubSub struct {
	PublishFn   func(ctx context.Context, m model.Message) error
	SubscribeFn func(ctx context.Context) (<-chan model.Message, func() error, error)
}

var _ pubsub.PubSub = (*MockPubSub)(nil)

func (m *MockPubSub) Publish(ctx context.Context, msg model.Message) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, msg)
	}
	return nil
}
func (m *MockPubSub) Subscribe(ctx context.Context) (<-chan model.Message, func() error, error) {
	if m.SubscribeFn != nil {
		return m.SubscribeFn(ctx)
	}
	ch := make(chan model.Message)
	close(ch)
	return ch, func() error { return nil }, nil
}
func (m *MockPubSub) Close() error { return nil }
