package executor

import (
	"context"
	"fmt"
	"time"
)

func Execute(ctx context.Context,payload string) error {

	// 
	fmt.Println("work done:",payload)

	select {
	case <-time.After(1 *time.Second):
		return nil 
	case <-ctx.Done():
		return ctx.Err()
	}
}