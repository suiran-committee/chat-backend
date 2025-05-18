package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients     = make(map[*websocket.Conn]bool)
	broadcast   = make(chan Message)
	mu          sync.Mutex
	ctx         = context.Background()
	rdb         *redis.Client
	db          *sql.DB
	frontOrigin string
)

func main() {
	frontOrigin = getenv("FRONTEND_ORIGIN", "*")

	rdb = redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "localhost:6379")})

	var err error
	db, err = sql.Open("postgres", dbConnStr())
	if err != nil {
		log.Fatal(err)
	}

	go subscribeRedis()
	go handleBroadcast()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/messages", withCORS(handleHistory))

	log.Println("Listening on :8443 (HTTPS/WSS)")
	err = http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
	log.Fatal(err)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
		_, _ = db.Exec("INSERT INTO messages (username, content) VALUES ($1, $2)", msg.Username, msg.Content)
		jsonData, _ := json.Marshal(msg)
		rdb.Publish(ctx, "chat", jsonData)
	}
}

func subscribeRedis() {
	sub := rdb.Subscribe(ctx, "chat")
	ch := sub.Channel()
	for msg := range ch {
		var m Message
		_ = json.Unmarshal([]byte(msg.Payload), &m)
		broadcast <- m
	}
}

func handleBroadcast() {
	for msg := range broadcast {
		mu.Lock()
		for conn := range clients {
			conn.WriteJSON(msg)
		}
		mu.Unlock()
	}
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT username, content FROM messages ORDER BY id DESC LIMIT 100")
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		_ = rows.Scan(&m.Username, &m.Content)
		messages = append([]Message{m}, messages...)
	}
	json.NewEncoder(w).Encode(messages)
}

func dbConnStr() string {
	host := getenv("DB_HOST", "localhost")
	user := getenv("DB_USER", "chat")
	pass := getenv("DB_PASSWORD", "chat")
	name := getenv("DB_NAME", "chat")
	return "host=" + host + " user=" + user + " password=" + pass + " dbname=" + name + " sslmode=disable"
}

func getenv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}
