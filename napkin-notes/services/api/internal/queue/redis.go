// Package queue provides a Redis-backed job queue for font generation.
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// FontJob represents a font generation job to be processed by the worker.
type FontJob struct {
	FontID           string `json:"font_id"`
	UserID           string `json:"user_id"`
	TemplateScanPath string `json:"template_scan_path"`
	OutputPath       string `json:"output_path"`
}

// RedisQueue is a Redis-backed queue for font generation jobs.
type RedisQueue struct {
	client    *redis.Client
	queueName string
}

// NewRedisQueue creates a new RedisQueue connected to the given Redis URL.
func NewRedisQueue(redisURL, queueName string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	return &RedisQueue{
		client:    client,
		queueName: queueName,
	}, nil
}

// Enqueue adds a font generation job to the queue.
func (q *RedisQueue) Enqueue(ctx context.Context, job FontJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = q.client.LPush(ctx, q.queueName, data).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}
