package pubsub

import (
	"context"
	"encoding/json"

	"chat-backend/internal/model"

	"github.com/go-redis/redis/v8"
)

type PubSub interface {
	Publish(context.Context, model.Message) error
	Subscribe(context.Context) (<-chan model.Message, func() error, error)
	Close() error
}

type redisPubSub struct {
	rdb *redis.Client
}

const channel = "chat"

func NewRedis(addr string) PubSub {
	return &redisPubSub{
		rdb: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (r *redisPubSub) Publish(ctx context.Context, m model.Message) error {
	b, _ := json.Marshal(m)
	return r.rdb.Publish(ctx, channel, b).Err()
}

func (r *redisPubSub) Subscribe(ctx context.Context) (<-chan model.Message, func() error, error) {
	sub := r.rdb.Subscribe(ctx, channel)
	ch := make(chan model.Message)
	go func() {
		for msg := range sub.Channel() {
			var m model.Message
			_ = json.Unmarshal([]byte(msg.Payload), &m)
			ch <- m
		}
		close(ch)
	}()
	return ch, sub.Close, nil
}

func (r *redisPubSub) Close() error { return r.rdb.Close() }
