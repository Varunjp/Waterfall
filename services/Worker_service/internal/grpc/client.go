package grpcclient

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"sync/atomic"
	"time"
	pb "worker_service/internal/grpc/schedulerpb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Client struct {
	api pb.SchedulerClient
	conn *grpc.ClientConn
	log *zap.Logger

	stream pb.Scheduler_JobStreamClient

	appID string 
	workerID string 
	jobTypes []string 
	maxConcurrency int 

	activeJobs atomic.Int32

	sendQueue chan *pb.WorkerMessage
	JobQueue chan *pb.JobAssignment

	ctx context.Context
	cancel context.CancelFunc

	closeOnce sync.Once
}

func NewGrpcClient(addr string,log *zap.Logger) (*Client,error) {
	conn, err := grpc.NewClient(addr,grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	//conn, err := grpc.NewClient(addr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil,err 
	}

	api := pb.NewSchedulerClient(conn)
	ctx,cancel := context.WithCancel(context.Background())

	return &Client{
		api: api,
		conn: conn,
		log: log,

		sendQueue: make(chan *pb.WorkerMessage,100),
		JobQueue: make(chan *pb.JobAssignment,1000),

		ctx: ctx,
		cancel: cancel,
	},nil 
}

func (c *Client) Start (ctx context.Context, appID string, workerID string, jobTypes []string, maxConcurrency int) error {

	c.appID = appID
	c.workerID = workerID
	c.jobTypes = jobTypes
	c.maxConcurrency = maxConcurrency

	return c.connectionLoop(ctx)
}

func (c *Client) connectionLoop(ctx context.Context) error{
	backoff := time.Second

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		default:
		}
		err := c.connect(ctx)

		if err != nil {
			
			if ctx.Err() != nil || errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
				c.log.Info("grpc stream stopped gracefully")
				return nil
			}

			c.log.Error("grpc disconnected",zap.Error(err))
			timer := time.NewTimer(backoff)	

			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <- timer.C:
			
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}

			continue
		}
		backoff = time.Second
	}
}

func (c *Client) connect(ctx context.Context) error {

	stream, err :=
		c.api.JobStream(ctx)

	if err != nil {
		return err
	}

	c.stream = stream

	err = c.register()
	if err != nil {
		return err
	}

	writeCtx,cancel := context.WithCancel(ctx)
	defer cancel()

	go c.writePump(writeCtx)

	go c.heartbeatLoop(writeCtx)

	for {

		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch payload :=
		msg.Payload.(type) {

		case *pb.SchedulerMessage_Job:

			select {
			case <- ctx.Done():
				return ctx.Err()
			case c.JobQueue <- payload.Job:
			default:
				c.log.Warn("job queue full")
			}
		}
	}
}

func (c *Client) register() error {

	return c.send(
		&pb.WorkerMessage{
			Payload:
			&pb.WorkerMessage_Register{
				Register:
				&pb.WorkerRegister{
					AppId: c.appID,
					WorkerId:
						c.workerID,
					JobTypes:
						c.jobTypes,
					MaxConcurrency:
						int32(
							c.maxConcurrency,
						),
				},
			},
		},
	)
}

func (c *Client) heartbeatLoop(ctx context.Context) {

	ticker :=
		time.NewTicker(
			2 * time.Second,
		)

	defer ticker.Stop()

	for {

		select {

		case <- ctx.Done():
			return

		case <-ticker.C:

			_ = c.send(
				&pb.WorkerMessage{
					Payload:
					&pb.WorkerMessage_Heartbeat{
						Heartbeat:
						&pb.WorkerHeartbeat{
							ActiveJobs:
							c.activeJobs.Load(),
							Timestamp:
							time.Now().Unix(),
						},
					},
				},
			)
		}
	}
}

func (c *Client) writePump(ctx context.Context) {

	for {

		select {

		case <-ctx.Done():
			return

		case msg := <-c.sendQueue:

			if c.stream == nil {
				return
			}

			err :=
				c.stream.Send(msg)

			if err != nil {
				return
			}
		}
	}
}

func (c *Client) send(
	msg *pb.WorkerMessage,
) error {

	select {

	case <-c.ctx.Done():
		return c.ctx.Err()

	case c.sendQueue <- msg:
		return nil

	case <-time.After(
		3 * time.Second,
	):
		return errors.New(
			"send timeout",
		)
	}
}

func (c *Client) ReportResult(
	success bool,
	jobID string,
	errMsg string,
	retry int,
) {

	status :=
		pb.JobResultStatus_JOB_RESULT_SUCCESS

	if !success {
		status =
			pb.JobResultStatus_JOB_RESULT_FAILED
	}

	_ = c.send(
		&pb.WorkerMessage{
			Payload:
			&pb.WorkerMessage_Result{
				Result:
				&pb.JobExecutionResult{
					JobId: jobID,
					Status: status,
					ErrorMessage:
						errMsg,
					Retry:
						int32(retry),
				},
			},
		},
	)
}

func (c *Client) Progress(
	jobID string,
	progress int64,
) {
	switch progress {
	case 0:
		c.activeJobs.Add(1)
	case 100:
		c.activeJobs.Add(-1)
	}
	_ = c.send(
		&pb.WorkerMessage{
			Payload:
			&pb.WorkerMessage_Progress{
				Progress:
				&pb.JobProgress{
					JobId: jobID,
					Progress:
						progress,
					Timestamp:
						time.Now().
						Unix(),
				},
			},
		},
	)
}

func (c *Client) Close() {

	c.closeOnce.Do(func() {
		c.cancel()

		close(c.sendQueue)

		_ = c.conn.Close()
	})
}
// func (c *Client) Heartbeat(ctx context.Context, jobID, appID, workerID string, progress int64) {
// 	if _, err := c.api.Heartbeat(ctx, &pb.HeartbeatRequest{
// 		JobId:     jobID,
// 		AppId:     appID,
// 		WorkerId:  workerID,
// 		Progress:  progress,
// 		Timestamp: time.Now().Unix(),
// 	}); err != nil {
// 		return
// 	}
// }

// func (c *Client) ReportResult(ctx context.Context, jobID, appID, workerID string, success bool, errMsg string, retry int, manual_retry int) {
// 	status := pb.JobResultStatus_JOB_RESULT_SUCCESS
// 	if !success {
// 		status = pb.JobResultStatus_JOB_RESULT_FAILED
// 	}

// 	if _, err := c.api.ReportResult(ctx, &pb.JobResultRequest{
// 		JobId:        jobID,
// 		AppId:        appID,
// 		WorkerId:     workerID,
// 		Status:       status,
// 		ErrorMessage: errMsg,
// 		Retry:        int32(retry),
// 		ManualRetry:  int32(manual_retry),
// 		Timestamp:    time.Now().Unix(),
// 	}); err != nil {
// 		return
// 	}
// }

// func (c *Client) RegisterWorker(ctx context.Context, appID, workerID string, jobTypes []string, maxConcurrency int) {
// 	_, _ = c.api.RegisterWorker(ctx, &pb.RegisterWorkerRequest{
// 		AppId:          appID,
// 		WorkerId:       workerID,
// 		JobTypes:       jobTypes,
// 		MaxConcurrency: int32(maxConcurrency),
// 		Timestamp:      time.Now().Unix(),
// 	})
// }

// func (c *Client) WorkerHeartbeat(ctx context.Context, appID, workerID string, jobTypes []string, activeJobs int, maxConcurrency int) {
// 	_, _ = c.api.WorkerHeartbeat(ctx, &pb.WorkerHeartbeatRequest{
// 		AppId:          appID,
// 		WorkerId:       workerID,
// 		JobTypes:       jobTypes,
// 		ActiveJobs:     int32(activeJobs),
// 		MaxConcurrency: int32(maxConcurrency),
// 		Timestamp:      time.Now().Unix(),
// 	})
// }

// func (c *Client) UnregisterWorker(ctx context.Context, appID, workerID string) {
// 	_, _ = c.api.UnregisterWorker(ctx, &pb.UnregisterWorkerRequest{
// 		AppId:     appID,
// 		WorkerId:  workerID,
// 		Timestamp: time.Now().Unix(),
// 	})
// }
