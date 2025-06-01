package database

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    Client *redis.Client
}

func NewRedisConnection(redisURL string) (*RedisClient, error) {
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }

    client := redis.NewClient(opt)

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = client.Ping(ctx).Result()
    if err != nil {
        return nil, err
    }

    return &RedisClient{Client: client}, nil
}

func (r *RedisClient) Close() error {
    return r.Client.Close()
}