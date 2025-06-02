package pubsub

import (
	"context"
	"testing"
	"time"

	"chat-backend/internal/model"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestRedisPubSub_PublishSubscribe(t *testing.T) {
	srv := miniredis.RunT(t)
	ps := NewRedis(srv.Addr())

	ctx := context.Background()
	subCh, closeFn, err := ps.Subscribe(ctx)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer closeFn()

	want := model.Message{Username: "charlie", Content: "ping"}
	if err := ps.Publish(ctx, want); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case got := <-subCh:
		if got != want {
			t.Fatalf("mismatch: got %+v, want %+v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}
