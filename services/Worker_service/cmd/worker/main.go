package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
	"worker_service/internal/client"
	"worker_service/internal/config"
	"worker_service/internal/executor"
	pb "worker_service/internal/grpc/schedulerpb"
	"worker_service/internal/heartbeat"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}

	cfg := config.Load()

	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(){
		sig := make(chan os.Signal,1)
		signal.Notify(sig,os.Interrupt)
		<-sig
		cancel()
	}()

	scheduler,err := client.NewSchedulerClient(cfg.SchedulerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer scheduler.Close()

	_,err = scheduler.RegisterWorker(ctx,&pb.RegisterWorkerRequest{
		WorkerId: cfg.WorkerID,
		AppId: cfg.AppID,
		Capabilities: cfg.Capabilities,
		MaxConcurrency: 6,
	})

	if err != nil {
		log.Fatal("register failed:",err)
	}

	go heartbeat.Start(
		ctx,
		scheduler,
		cfg.WorkerID,
		time.Duration(cfg.HeartbeatSec) * time.Second,
	)

	for {
		select {
		case <- ctx.Done():
			log.Println("worker shutting down")
			return 
		default:
			resp,err := scheduler.WaitForJob(ctx,&pb.WaitForJobRequest{
				WorkerId: cfg.WorkerID,
				AppId: cfg.AppID,
				Capabilities: cfg.Capabilities,
			})

			if err != nil {
				log.Println("wait error :",err)
				continue
			}

			job := resp.Job
			exec,ok := executor.Registry[job.JobType]
			if !ok {
				log.Println("no executor for job:",job.JobType)
				continue
			}

			err = exec(ctx,job.Payload)

			if err != nil {
				_,_ = scheduler.CompleteJob(ctx,&pb.CompleteJobRequest{
					JobId: job.JobId,
					AppId: job.AppId,
					JobType: job.JobType,
					WorkerId: cfg.WorkerID,
					StreamId: resp.StreamId,
					Success: false,
					ErrorMessage: func()string{
						if err != nil{
							return err.Error()
						}
						return ""
					}(),
				})
				continue 
			}

			_,_ = scheduler.CompleteJob(ctx,&pb.CompleteJobRequest{
				JobId: job.JobId,
				AppId: job.AppId,
				JobType: job.JobType,
				WorkerId: cfg.WorkerID,
				StreamId: resp.StreamId,
				Success: true,
				ErrorMessage: func()string{
					if err != nil{
						return err.Error()
					}
					return ""
				}(),
			})
		}
	}
}