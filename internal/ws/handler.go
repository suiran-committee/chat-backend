package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"chat-backend/internal/model"
	"chat-backend/internal/pubsub"
	"chat-backend/internal/storage"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Hub          *Hub
	Store        storage.HistoryStore
	PubSub       pubsub.PubSub
	FrontendOrig string
	upgrader     websocket.Upgrader
}

func NewHandler(h *Hub, st storage.HistoryStore, ps pubsub.PubSub, origin string) *Handler {
	return &Handler{
		Hub:          h,
		Store:        st,
		PubSub:       ps,
		FrontendOrig: origin,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade: %v", err)
		return
	}
	h.Hub.Register(conn)
	defer func() { h.Hub.Unregister(conn); conn.Close() }()

	for {
		var m model.Message
		if err := conn.ReadJSON(&m); err != nil {
			log.Printf("[ws] read: %v", err)
			return
		}
		if err := h.Store.Save(r.Context(), m); err != nil {
			log.Printf("[db] save: %v", err)
		}
		if err := h.PubSub.Publish(r.Context(), m); err != nil {
			log.Printf("[redis] publish: %v", err)
		}
	}
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	withCORS(h.FrontendOrig, func(w http.ResponseWriter, r *http.Request) {
		msgs, err := h.Store.Last(r.Context(), 100)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(msgs)
	})(w, r)
}

func withCORS(origin string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (h *Handler) StartRedisFanIn(ctx context.Context) error {
	subCh, closeFn, err := h.PubSub.Subscribe(ctx)
	if err != nil {
		return err
	}
	go func() {
		defer closeFn()
		for m := range subCh {
			h.Hub.Send(m)
		}
	}()
	return nil
}
