package grpcserver

import (
	"context"
	schedulerpb "scheduler_service/internal/grpc/schedulerpb"
	"sync"
	"sync/atomic"
)

type WorkerConnection struct {
	AppID	string 
	WorkerID string 

	JobTypes	[]string 
	MaxConcurrency	int32 
	
	activeJobs 	 atomic.Int32 

	Stream schedulerpb.Scheduler_JobStreamServer
	SendQueue chan *schedulerpb.SchedulerMessage

	Ctx context.Context
	Cancel context.CancelFunc

	mu sync.Mutex
	closeOnce sync.Once
}

func (w *WorkerConnection) CanTakeJob() bool {
	return w.activeJobs.Load() < w.MaxConcurrency
}

func (w *WorkerConnection) ActiveJobs() int32 {
	return w.activeJobs.Load()
}

func (w *WorkerConnection) IncrementJobs() {
	w.activeJobs.Add(1)
}

func (w *WorkerConnection) DecrementJobs() {
	w.activeJobs.Add(-1)
}

func (w *WorkerConnection) WritePump() {
	defer w.Cancel()

	for {
		select {
		case <-w.Ctx.Done():
			return 
		case msg,ok := <-w.SendQueue:
			if !ok {
				return 
			}

			err := w.Stream.Send(msg)
			if err != nil {
				return 
			}
		}
	}
}

func (w *WorkerConnection) Close() {
	w.closeOnce.Do(func(){
		w.Cancel()
		close(w.SendQueue)
	})
}

type WorkerManger struct {
	mu sync.RWMutex
	workers map[string]*WorkerConnection
}

func NewWorkerManager() *WorkerManger {
	return &WorkerManger{
		workers: make(map[string]*WorkerConnection),
	}
}

func (wm *WorkerManger) Register(worker *WorkerConnection) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.workers[worker.WorkerID] = worker
}

func (wm *WorkerManger) Remove(workerID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	delete(wm.workers,workerID)
}

func (wm *WorkerManger) FindAvailableWorker(appID string, jobType string) *WorkerConnection {

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var best *WorkerConnection

	for _,worker := range wm.workers {

		if worker.AppID != appID {
			continue
		}

		if !contains(worker.JobTypes,jobType) {
			continue
		}

		if !worker.CanTakeJob() {
			continue
		}

		if best == nil || worker.ActiveJobs() < best.ActiveJobs(){
			best = worker
		}

	}

	return best
}

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}