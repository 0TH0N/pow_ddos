package storage

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"powserver/config"
	"strconv"
	"time"
)

type TaskStorage struct {
	rdb *redis.Client
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.RedisUrl,
			Password: config.RedisPass,
			DB:       config.RedisDb,
		}),
	}
}

func (ts *TaskStorage) Add(ctx context.Context, task string) error {
	err := ts.rdb.Set(ctx, task, false, time.Minute*5).Err()
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskStorage) Get(ctx context.Context, task string) (bool, bool, error) {
	val, err := ts.rdb.Get(ctx, task).Result()
	if errors.Is(err, redis.Nil) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return false, false, err
	}

	return boolValue, true, nil
}

func (ts *TaskStorage) Mark(ctx context.Context, task string) error {
	err := ts.rdb.Set(ctx, task, true, time.Minute*5).Err()
	if err != nil {
		return err
	}

	return nil
}
