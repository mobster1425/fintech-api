package worker

import (
	"context"

	"github.com/hibiken/asynq"
)



// TaskDistributor interface defines methods for distributing tasks
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error
}

// RedisTaskDistributor struct implements the TaskDistributor interface using Redis and Asynq
type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor initializes and returns a new RedisTaskDistributor
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}