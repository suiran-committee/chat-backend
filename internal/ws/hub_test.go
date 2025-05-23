package ws

import (
	"context"
	"strings"
	"testing"
	"time"

	"chat-backend/internal/model"
	"chat-backend/internal/testutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/websocket"
)

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub(4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewHandler(
		hub,
		&testutil.MockStore{},
		&testutil.MockPubSub{},
		"*",
	)
	ts := httptest.NewServer(http.HandlerFunc(handler.WebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	c1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close()

	c2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.Close()

	want := model.Message{Username: "bob", Content: "hello"}
	hub.Send(want)

	read := func(c *websocket.Conn) model.Message {
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		var m model.Message
		if err := c.ReadJSON(&m); err != nil {
			t.Fatalf("read: %v", err)
		}
		return m
	}
	if got := read(c1); got != want {
		t.Fatalf("c1 got %+v, want %+v", got, want)
	}
	if got := read(c2); got != want {
		t.Fatalf("c2 got %+v, want %+v", got, want)
	}
}
