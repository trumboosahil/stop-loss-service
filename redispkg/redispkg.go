package redispkg

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	Publish(channel, message string) error
	Subscribe(channel string) <-chan string
	ZAdd(key string, member string, score float64) error
	ZPopMinBatch(key string, count int64) ([]redis.Z, error)
}

type RedisClientImpl struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(addr string) *RedisClientImpl {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RedisClientImpl{client: client, ctx: context.Background()}
}

// Publish sends a message to a Redis Pub/Sub channel
func (r *RedisClientImpl) Publish(channel, message string) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

// Subscribe listens for messages on a Redis Pub/Sub channel
func (r *RedisClientImpl) Subscribe(channel string) <-chan string {
	pubsub := r.client.Subscribe(r.ctx, channel)
	ch := make(chan string)

	go func() {
		defer close(ch)
		for msg := range pubsub.Channel() {
			ch <- msg.Payload
		}
	}()
	return ch
}

// ZAdd adds a member to a sorted set with a given score
func (r *RedisClientImpl) ZAdd(key string, member string, score float64) error {
	return r.client.ZAdd(r.ctx, key, &redis.Z{Member: member, Score: score}).Err()
}

// ZPopMinBatch retrieves and removes multiple members with the lowest scores from a sorted set
func (r *RedisClientImpl) ZPopMinBatch(key string, count int64) ([]redis.Z, error) {
	result, err := r.client.ZPopMin(r.ctx, key, count).Result()
	if err != nil || len(result) == 0 {
		return nil, err
	}
	return result, nil
}

// Close closes the Redis client connection
func (r *RedisClientImpl) Close() error {
	return r.client.Close()
}
