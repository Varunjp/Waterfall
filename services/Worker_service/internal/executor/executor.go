package executor

import (
	"context"
	"errors"
	"fmt"
	"time"
	pb "worker_service/internal/grpc/schedulerpb"
)

func ExecuteJob(ctx context.Context,job *pb.JobAssignment) error {

	// 
	// fmt.Println("Work started")
	// time.Sleep(time.Second * 3)
	// // err := errors.New("TESTING error")

	// // return err 
	// fmt.Println("work done:",payload)

	// select {
	// case <-time.After(1 *time.Second):
	// 	return nil 
	// case <-ctx.Done():
	// 	return ctx.Err()
	// }
	
	fmt.Println("Work started")
	time.Sleep(time.Second * 3)

	switch job.JobType {

	case "email":
		fmt.Println("worke done: ",job.Payload)

		return nil 

	case "notification":
		return nil 

	default:
		return errors.New(
			"unsupported job",
		)
	}
}