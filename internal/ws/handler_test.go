package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat-backend/internal/model"
	"chat-backend/internal/testutil"
)

func TestHandler_History(t *testing.T) {
	want := []model.Message{{Username: "alice", Content: "hi"}}

	h := NewHandler(
		NewHub(4),
		&testutil.MockStore{
			LastFn: func(_ context.Context, n int) ([]model.Message, error) { return want, nil },
		},
		&testutil.MockPubSub{},
		"http://example.com",
	)

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	rec := httptest.NewRecorder()

	h.History(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("CORS header mismatch: %s", got)
	}
	var msgs []model.Message
	if err := json.NewDecoder(rec.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != want[0] {
		t.Fatalf("body mismatch: %+v", msgs)
	}
}
