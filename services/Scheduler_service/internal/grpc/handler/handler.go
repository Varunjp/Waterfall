package handler

import (
	"context"
	"scheduler_service/internal/domain"
	pb "scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/usecase"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchedulerHandler struct {
	pb.UnimplementedSchedulerServiceServer
	uc *usecase.SchedulerUsecase
}

func NewSchedulerHandler(uc *usecase.SchedulerUsecase) *SchedulerHandler {
	return &SchedulerHandler{uc : uc}
}

func (h *SchedulerHandler) RegisterWorker(ctx context.Context, req *pb.RegisterWorkerRequest) (*pb.RegisterWorkerResponse,error) {
	err := h.uc.RegisterWorker(domain.Worker{
		WorkerID: req.WorkerId,
		AppID: req.AppId,
		Capabilities: req.Capabilities,
		Concurrency: int(req.MaxConcurrency),
		LastSeen: time.Now(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal,"register worker failed: %v",err)
	}

	return &pb.RegisterWorkerResponse{
		Accepted: true,
	},nil 
}

func (h *SchedulerHandler) WorkerHeartbeat(ctx context.Context, req *pb.WorkerHeartbeatRequest) (*pb.Empty,error) {
	if err := h.uc.WorkerHeartbeat(req.WorkerId); err != nil {
		return nil, status.Errorf(codes.Internal,"heartbeat failed")
	}

	return &pb.Empty{},nil 
}

// func (h *SchedulerHandler) PollJob(ctx context.Context, req *pb.PollJobRequest) (*pb.PollJobResponse,error) {
// 	job,err := h.uc.AssignJob(
// 		ctx,
// 		req.WorkerId,
// 		req.AppId,
// 		req.Capabilities,
// 	)
	
// 	if err != nil {
// 		if errors.Is(err, usecase.ErrNoJobAvailable) {
// 			return &pb.PollJobResponse{Found: false},nil 
// 		}
// 		if errors.Is(err, usecase.ErrQuotaExceeded) {
// 			return nil, status.Errorf(codes.ResourceExhausted,"quota exceeded")
// 		}
// 		return nil, status.Errorf(codes.Internal, "%s", err.Error())
// 	}

// 	return &pb.PollJobResponse{
// 		Found: true,
// 		Job: &pb.Job{
// 			JobId: job.JobID,
// 			AppId: job.AppID,
// 			JobType: job.Type,
// 			Payload: []byte(job.Payload),
// 			Attempt: int32(job.Retry),
// 			MaxAttempts: int32(job.MaxRetries),
// 		},
// 	},nil 
// }

func (h *SchedulerHandler) WaitForJob(ctx context.Context,req *pb.WaitForJobRequest,)(*pb.WaitForJobResponse,error) {

	job,msgID, err := h.uc.ConsumeJobForWorker(
		ctx,
		domain.Worker{
			WorkerID: req.WorkerId,
			AppID: req.AppId,
			Capabilities: req.Capabilities,
		},
	)

	if err != nil || job == nil {
		return &pb.WaitForJobResponse{Found: false},nil 
	}

	return &pb.WaitForJobResponse{
		Found:true,
		StreamId: msgID,
		Job: &pb.Job{
			JobId: job.JobID,
			AppId: job.AppID,
			JobType: job.Type,
			Payload: []byte(job.Payload),
		},
	},nil 
}

func (h *SchedulerHandler) JobHeartbeat(ctx context.Context,req *pb.JobHeartbeatRequest)(*pb.Empty,error) {
	err := h.uc.JobHeartbeat(
		ctx,
		req.JobId,
		req.WorkerId,
		req.Message,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal,"job heartbeat failed")
	}

	return &pb.Empty{},nil 
}

func (h *SchedulerHandler) CompleteJob(ctx context.Context, req *pb.CompleteJobRequest) (*pb.Empty,error) {

	success := req.Success

	job := domain.Job{
		JobID: req.JobId,
		AppID: req.AppId,
		Type: req.JobType,
	}

	err := h.uc.CompleteStreamJob(
		ctx,
		job,
		req.StreamId,
		success,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal,"complete job failed")
	}

	return &pb.Empty{},nil 
}

func (h *SchedulerHandler) PushJobLog(ctx context.Context, req *pb.PushJobLogRequest)(*pb.Empty,error) {
	err := h.uc.PushJobLog(
		ctx,
		req.JobId,
		req.WorkerId,
		req.Status,
		req.ErrorMessage,
		int(req.Attempt),
		time.Unix(req.TimestampUnix,0), 
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal,"log push failed")
	}

	return &pb.Empty{},nil 
}

