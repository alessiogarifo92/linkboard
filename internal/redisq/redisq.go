// Package redisq implementa la coda dei click su una lista Redis (LPUSH / BRPOP).
package redisq

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultKey è la lista Redis dei click.
const DefaultKey = "linkboard:clicks"

type Queue struct {
	Client *redis.Client
	Key    string
}

func New(addr string) *Queue {
	return &Queue{Client: redis.NewClient(&redis.Options{Addr: addr}), Key: DefaultKey}
}

func (q *Queue) Push(ctx context.Context, code string) error {
	return q.Client.LPush(ctx, q.Key, code).Err()
}

// Pop attende fino a timeout un click; restituisce "", nil se la coda resta vuota.
func (q *Queue) Pop(ctx context.Context, timeout time.Duration) (string, error) {
	res, err := q.Client.BRPop(ctx, timeout, q.Key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return res[1], nil // res[0] è il nome della chiave
}

func (q *Queue) Ping(ctx context.Context) error { return q.Client.Ping(ctx).Err() }
