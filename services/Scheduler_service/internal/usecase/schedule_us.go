package usecase

import (
	"context"
	"errors"
	"fmt"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/repository"
	"time"
)

var (
	ErrNoJobAvailable = errors.New("no job available")
	ErrQuotaExceeded  = errors.New("app quota exceeded")
)

type SchedulerUsecase struct {
	jobStore repository.JobStore
	workerReg repository.WorkerRegistry
	quotaRepo repository.AdminQuotaRepository
	eventQueue repository.JobEventQueue
	metrics   *metrics.SchedulerMetrics
	logQueue  repository.JobLogQueue
	stallTimeout time.Duration
}

func NewSchedulerUsecase(
	jobStore repository.JobStore,
	workerReg repository.WorkerRegistry,
	quotaRepo repository.AdminQuotaRepository,
	eventQueue repository.JobEventQueue,
	metrics *metrics.SchedulerMetrics,
	logqueu repository.JobLogQueue,
) *SchedulerUsecase {
	return &SchedulerUsecase{
		jobStore: jobStore,
		workerReg: workerReg,
		quotaRepo: quotaRepo,
		eventQueue: eventQueue,
		metrics: metrics,
		logQueue: logqueu,
		stallTimeout: 60 * time.Second,
	}
}

// call for grpc PollJob

// func (s *SchedulerUsecase) AssignJob(
// 	ctx context.Context,
// 	workerID string,
// 	appID string, 
// 	capabilities []string,
// )(*domain.Job,error) {
	
// 	ok,err := s.quotaRepo.CanStart(appID)
// 	if err != nil || !ok {
// 		return nil, ErrQuotaExceeded
// 	}

// 	for _, jobType := range capabilities {
// 		job,err := s.jobStore.PollJob(appID,jobType,workerID)

// 		if err != nil || job == nil {
// 			continue
// 		}

// 		if _,err := s.workerReg.Acquire(appID, jobType); err != nil {
// 			_ = s.jobStore.Requeue(job)
// 			return nil,err 
// 		}

// 		_ = s.quotaRepo.Increment(appID)

// 		_ = s.eventQueue.JobRunning(job.JobID,workerID)

// 		s.metrics.JobsAssigned.Inc()
// 		s.metrics.RunningJobs.Inc()

// 		return job,nil 
// 	}

// 	return nil, ErrNoJobAvailable
// }

func (s *SchedulerUsecase) ConsumeJobForWorker(ctx context.Context, worker domain.Worker) (*domain.Job,string,error) {

	ok,err := s.quotaRepo.CanStart(worker.AppID)
	if err != nil || !ok {
		return nil,"",ErrQuotaExceeded
	}

	for _,jobType := range worker.Capabilities {

		_ = s.jobStore.EnsureGroup(worker.AppID,jobType,"workers")

		job,msgID,err := s.jobStore.ReadJob(
			ctx,
			worker.AppID,
			jobType,
			"workers",
			worker.WorkerID,
		)

		if job == nil || err != nil {
			//delete
			fmt.Println("check error scheduler :",err)
			fmt.Println("check jobs schduler:",job)
			continue 
		}

		_ = s.quotaRepo.Increment(worker.AppID)
		s.metrics.JobsAssigned.Inc()

		return job,msgID,nil 
	}

	return nil,"",ErrNoJobAvailable
}

func (s *SchedulerUsecase) CompleteStreamJob(ctx context.Context,job domain.Job,streamID string, success bool) error {

	err := s.jobStore.AckJob(job.AppID,job.Type,"workers",streamID)
	if err != nil {
		return err
	}
	_ = s.quotaRepo.Decrement(job.AppID)

	if success {
		s.metrics.JobsSuccess.Inc()
		return s.eventQueue.JobSucceeded(job.JobID)
	}

	s.metrics.JobsFailed.Inc()
	return s.eventQueue.JobFailed(job.JobID,"execution failed")
}

func (s *SchedulerUsecase) RegisterWorker(worker domain.Worker) error {
	worker.LastSeen = time.Now()
	worker.ActiveJobs = 0

	return s.workerReg.Register(worker)
}

// Worker heartbeat

func (s *SchedulerUsecase) WorkerHeartbeat(workerID string) error {
	return s.workerReg.Heartbeat(workerID)
}

func (s *SchedulerUsecase) JobHeartbeat(
	ctx context.Context,
	jobID string,
	workerID string,
	message string,
) error {
	err := s.jobStore.RefershRunning(jobID)
	if err != nil {
		return err 
	}

	// need to implement logic for job update in db

	return s.eventQueue.JobRunning(jobID,workerID)
}

// job completion

// func (s *SchedulerUsecase) CompleteJob(
// 	ctx context.Context,
// 	jobID string,
// 	workerID string,
// 	appID string,
// 	success bool,
// 	errorMsg string,
// ) error {

// 	_ = s.jobStore.DeleteRunning(jobID)
// 	_ = s.workerReg.Release(workerID)
// 	//_ = s.quotaRepo.Decrement(appID)

// 	s.metrics.RunningJobs.Dec()

// 	if success {
// 		s.metrics.JobsSuccess.Inc()
// 		return s.eventQueue.JobSucceeded(jobID)
// 	}

// 	s.metrics.JobsFailed.Inc()
// 	return s.eventQueue.JobFailed(jobID,errorMsg)
// }

func (s *SchedulerUsecase) StartStalledJobReaper(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <- ctx.Done():
			return 
		case <- ticker.C:
			s.reapStalledJobs()
		}
	}
}

func (s *SchedulerUsecase) reapStalledJobs() {
	stalledJobs,err := s.jobStore.FindStalled(s.stallTimeout)
	if err != nil {
		return 
	}

	for _,job := range stalledJobs {
		_ = s.jobStore.DeleteRunning(job.JobID)
		_ = s.workerReg.Release(job.WorkerID)
		_ = s.quotaRepo.Decrement(job.AppID)

		s.metrics.JobsFailed.Inc()
		s.metrics.RunningJobs.Dec()

		_ = s.eventQueue.JobFailed(job.JobID,"worker heartbeat timeout")
	}
}

// Logs

func (s *SchedulerUsecase) PushJobLog(
	ctx context.Context,
	jobID string,
	workerID string,
	status string,
	errorMsg string,
	attempt int,
	logtime time.Time, 
) error {
	return s.logQueue.Push(jobID,workerID,status,errorMsg,attempt,logtime)
}