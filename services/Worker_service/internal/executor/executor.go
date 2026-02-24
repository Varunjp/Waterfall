package executor

import (
	"context"
	"fmt"
	"time"
)

func Execute(ctx context.Context,payload string) error {

	// 
	fmt.Println("Work started")
	time.Sleep(time.Second * 3)
	// err := errors.New("TESTING error")

	// return err 
	fmt.Println("work done:",payload)

	select {
	case <-time.After(1 *time.Second):
		return nil 
	case <-ctx.Done():
		return ctx.Err()
	}
}