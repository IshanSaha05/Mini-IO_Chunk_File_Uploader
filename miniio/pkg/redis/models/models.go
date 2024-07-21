package models

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client        *redis.Client
	Ctx           context.Context
	Once          sync.Once
	Cancel        context.CancelFunc
	ExpireChannel *redis.PubSub
	Wg            sync.WaitGroup
}
