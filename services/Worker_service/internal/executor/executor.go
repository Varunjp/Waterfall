package executor

import (
	"context"
)

type JobExecutor func(ctx context.Context, payload []byte) error

var Registry = map[string]JobExecutor{}

func Register(jobType string, fn JobExecutor) {
	Registry[jobType] = fn
}