package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"scheduler_service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	workerFreshFor = 20 * time.Second
	workerStaleFor = time.Minute
	stateTTL       = 24 * time.Hour
)

type Store struct {
	redis *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{redis: rdb}
}

func (s *Store) RecordQueuedJob(ctx context.Context, job domain.Job, ts time.Time) error {
	jobKey := s.jobKey(job.JobID)
	readyKey := s.readyKey(job.AppID, job.Type)

	jobFields := map[string]any{
		"app_id":      job.AppID,
		"job_type":    job.Type,
		"enqueued_at": ts.Unix(),
		"started_at":  0,
		"worker_id":   "",
	}

	pipe := s.redis.TxPipeline()
	pipe.SAdd(ctx, s.jobTypesKey(job.AppID), job.Type)
	pipe.ZAdd(ctx, readyKey, redis.Z{Score: float64(ts.Unix()), Member: job.JobID})
	pipe.HSet(ctx, jobKey, jobFields)
	pipe.Expire(ctx, jobKey, stateTTL)
	pipe.Expire(ctx, readyKey, stateTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) RecordWorkerRegistration(
	ctx context.Context,
	appID string,
	workerID string,
	jobTypes []string,
	maxConcurrency int,
	ts time.Time,
) error {
	return s.upsertWorker(ctx, appID, workerID, jobTypes, maxConcurrency, 0, ts, true)
}

func (s *Store) RecordWorkerHeartbeat(
	ctx context.Context,
	appID string,
	workerID string,
	jobTypes []string,
	maxConcurrency int,
	activeJobs int,
	ts time.Time,
) error {
	return s.upsertWorker(ctx, appID, workerID, jobTypes, maxConcurrency, activeJobs, ts, true)
}

func (s *Store) RecordWorkerSeen(
	ctx context.Context,
	appID string,
	workerID string,
	ts time.Time,
) error {
	return s.upsertWorker(ctx, appID, workerID, nil, 0, 0, ts, false)
}

func (s *Store) RecordJobStarted(
	ctx context.Context,
	appID string,
	workerID string,
	jobID string,
	ts time.Time,
) error {
	meta, err := s.redis.HGetAll(ctx, s.jobKey(jobID)).Result()
	if err != nil {
		return err
	}

	jobType := meta["job_type"]
	if appID == "" {
		appID = meta["app_id"]
	}
	if workerID == "" {
		workerID = meta["worker_id"]
	}
	if jobType == "" {
		return nil
	}

	runningKey := s.runningKey(appID, jobType)
	workerKey := s.workerKey(appID, workerID)

	pipe := s.redis.TxPipeline()
	pipe.ZRem(ctx, s.readyKey(appID, jobType), jobID)
	wasAdded := pipe.SAdd(ctx, runningKey, jobID)
	pipe.HSet(ctx, s.jobKey(jobID), map[string]any{
		"worker_id":  workerID,
		"started_at": ts.Unix(),
	})
	pipe.Expire(ctx, runningKey, stateTTL)
	pipe.Expire(ctx, s.jobKey(jobID), stateTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	if wasAdded.Val() > 0 {
		if err := s.touchWorkerCounter(ctx, workerKey, workerID, appID, ts, 1); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) RecordJobFinished(
	ctx context.Context,
	appID string,
	workerID string,
	jobID string,
	ts time.Time,
) error {
	return s.releaseJob(ctx, appID, workerID, jobID, ts)
}

func (s *Store) ReleaseJob(ctx context.Context, jobID string, ts time.Time) error {
	meta, err := s.redis.HGetAll(ctx, s.jobKey(jobID)).Result()
	if err != nil {
		return err
	}

	appID := meta["app_id"]
	workerID := meta["worker_id"]
	if appID == "" {
		return nil
	}

	return s.releaseJob(ctx, appID, workerID, jobID, ts)
}

func (s *Store) MarkWorkerOffline(ctx context.Context, appID, workerID string, ts time.Time) error {
	key := s.workerKey(appID, workerID)
	pipe := s.redis.TxPipeline()
	pipe.SAdd(ctx, s.workersKey(appID), workerID)
	pipe.HSet(ctx, key, map[string]any{
		"worker_id":    workerID,
		"app_id":       appID,
		"active_jobs":  0,
		"last_seen":    ts.Unix(),
		"worker_state": "offline",
	})
	pipe.Expire(ctx, key, stateTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) SnapshotTenant(ctx context.Context, appID string, now time.Time) (*domain.TenantRuntimeSnapshot, error) {
	jobTypes, err := s.redis.SMembers(ctx, s.jobTypesKey(appID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	workerIDs, err := s.redis.SMembers(ctx, s.workersKey(appID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	workers := make([]domain.WorkerSnapshot, 0, len(workerIDs))
	queues := make(map[string]*domain.QueueSnapshot, len(jobTypes))

	for _, jobType := range jobTypes {
		readyJobs, err := s.redis.ZCard(ctx, s.readyKey(appID, jobType)).Result()
		if err != nil {
			return nil, err
		}

		runningJobs, err := s.redis.SCard(ctx, s.runningKey(appID, jobType)).Result()
		if err != nil {
			return nil, err
		}

		var oldestReadyAge int64
		oldest, err := s.redis.ZRangeWithScores(ctx, s.readyKey(appID, jobType), 0, 0).Result()
		if err != nil {
			return nil, err
		}
		if len(oldest) > 0 {
			enqueuedAt := time.Unix(int64(oldest[0].Score), 0)
			if enqueuedAt.Before(now) {
				oldestReadyAge = int64(now.Sub(enqueuedAt).Seconds())
			}
		}

		queues[jobType] = &domain.QueueSnapshot{
			JobType:               jobType,
			ReadyJobs:             readyJobs,
			RunningJobs:           runningJobs,
			OldestReadyAgeSeconds: oldestReadyAge,
		}
	}

	snapshot := &domain.TenantRuntimeSnapshot{
		AppID:       appID,
		GeneratedAt: now,
	}

	for _, workerID := range workerIDs {
		meta, err := s.redis.HGetAll(ctx, s.workerKey(appID, workerID)).Result()
		if err != nil {
			return nil, err
		}
		if len(meta) == 0 {
			_ = s.redis.SRem(ctx, s.workersKey(appID), workerID).Err()
			continue
		}

		lastSeen := parseUnix(meta["last_seen"])
		activeJobs := parseInt(meta["active_jobs"])
		maxConcurrency := parseInt(meta["max_concurrency"])
		jobTypes := decodeJobTypes(meta["job_types"])
		status := deriveWorkerStatus(meta["worker_state"], now, lastSeen, activeJobs)

		worker := domain.WorkerSnapshot{
			WorkerID:       workerID,
			AppID:          appID,
			JobTypes:       jobTypes,
			ActiveJobs:     activeJobs,
			MaxConcurrency: maxConcurrency,
			LastSeen:       lastSeen,
			Status:         status,
		}
		workers = append(workers, worker)

		snapshot.TotalWorkers++
		switch status {
		case domain.WorkerStatusOnline:
			snapshot.OnlineWorkers++
		case domain.WorkerStatusBusy:
			snapshot.OnlineWorkers++
			snapshot.BusyWorkers++
		}

		for _, jobType := range jobTypes {
			queue, ok := queues[jobType]
			if !ok {
				queue = &domain.QueueSnapshot{JobType: jobType}
				queues[jobType] = queue
			}
			if status != domain.WorkerStatusOffline {
				queue.RegisteredWorkers++
			}
			if status == domain.WorkerStatusBusy {
				queue.BusyWorkers++
			}
		}
	}

	queueNames := make([]string, 0, len(queues))
	for jobType, queue := range queues {
		queueNames = append(queueNames, jobType)
		snapshot.TotalReadyJobs += queue.ReadyJobs
		snapshot.TotalRunningJobs += queue.RunningJobs
	}
	sort.Strings(queueNames)

	for _, jobType := range queueNames {
		snapshot.Queues = append(snapshot.Queues, *queues[jobType])
	}

	sort.Slice(workers, func(i, j int) bool {
		if workers[i].Status == workers[j].Status {
			return workers[i].WorkerID < workers[j].WorkerID
		}
		return workerOrder(workers[i].Status) < workerOrder(workers[j].Status)
	})
	snapshot.Workers = workers

	return snapshot, nil
}

func (s *Store) releaseJob(
	ctx context.Context,
	appID string,
	workerID string,
	jobID string,
	ts time.Time,
) error {
	meta, err := s.redis.HGetAll(ctx, s.jobKey(jobID)).Result()
	if err != nil {
		return err
	}

	if appID == "" {
		appID = meta["app_id"]
	}
	jobType := meta["job_type"]
	if jobType == "" {
		return nil
	}

	pipe := s.redis.TxPipeline()
	pipe.ZRem(ctx, s.readyKey(appID, jobType), jobID)
	removed := pipe.SRem(ctx, s.runningKey(appID, jobType), jobID)
	pipe.HSet(ctx, s.jobKey(jobID), map[string]any{
		"finished_at": ts.Unix(),
	})
	pipe.Expire(ctx, s.jobKey(jobID), stateTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	if removed.Val() > 0 && workerID != "" {
		if err := s.touchWorkerCounter(ctx, s.workerKey(appID, workerID), workerID, appID, ts, -1); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) upsertWorker(
	ctx context.Context,
	appID string,
	workerID string,
	jobTypes []string,
	maxConcurrency int,
	activeJobs int,
	ts time.Time,
	resetActive bool,
) error {
	fields := map[string]any{
		"worker_id":    workerID,
		"app_id":       appID,
		"last_seen":    ts.Unix(),
		"worker_state": "active",
	}

	if encoded := encodeJobTypes(jobTypes); encoded != "" {
		fields["job_types"] = encoded
	}
	if maxConcurrency > 0 {
		fields["max_concurrency"] = maxConcurrency
	}
	if resetActive || activeJobs > 0 {
		fields["active_jobs"] = activeJobs
	}

	key := s.workerKey(appID, workerID)

	pipe := s.redis.TxPipeline()
	pipe.SAdd(ctx, s.workersKey(appID), workerID)
	if len(jobTypes) > 0 {
		members := make([]any, 0, len(jobTypes))
		for _, jobType := range jobTypes {
			if strings.TrimSpace(jobType) == "" {
				continue
			}
			members = append(members, strings.TrimSpace(jobType))
		}
		if len(members) > 0 {
			pipe.SAdd(ctx, s.jobTypesKey(appID), members...)
		}
	}
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, stateTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) touchWorkerCounter(
	ctx context.Context,
	workerKey string,
	workerID string,
	appID string,
	ts time.Time,
	delta int64,
) error {
	pipe := s.redis.TxPipeline()
	pipe.SAdd(ctx, s.workersKey(appID), workerID)
	pipe.HSet(ctx, workerKey, map[string]any{
		"worker_id":    workerID,
		"app_id":       appID,
		"last_seen":    ts.Unix(),
		"worker_state": "active",
	})
	counter := pipe.HIncrBy(ctx, workerKey, "active_jobs", delta)
	pipe.Expire(ctx, workerKey, stateTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	if counter.Val() < 0 {
		return s.redis.HSet(ctx, workerKey, "active_jobs", 0).Err()
	}

	return nil
}

func (s *Store) workersKey(appID string) string {
	return fmt.Sprintf("tenant:runtime:workers:%s", appID)
}

func (s *Store) workerKey(appID string, workerID string) string {
	return fmt.Sprintf("tenant:runtime:worker:%s:%s", appID, workerID)
}

func (s *Store) jobTypesKey(appID string) string {
	return fmt.Sprintf("tenant:runtime:jobtypes:%s", appID)
}

func (s *Store) readyKey(appID string, jobType string) string {
	return fmt.Sprintf("tenant:runtime:ready:%s:%s", appID, jobType)
}

func (s *Store) runningKey(appID string, jobType string) string {
	return fmt.Sprintf("tenant:runtime:running:%s:%s", appID, jobType)
}

func (s *Store) jobKey(jobID string) string {
	return fmt.Sprintf("tenant:runtime:job:%s", jobID)
}

func deriveWorkerStatus(workerState string, now time.Time, lastSeen time.Time, activeJobs int) domain.WorkerRuntimeStatus {
	if workerState == "offline" {
		return domain.WorkerStatusOffline
	}
	if lastSeen.IsZero() {
		return domain.WorkerStatusOffline
	}

	age := now.Sub(lastSeen)
	switch {
	case age > workerStaleFor:
		return domain.WorkerStatusOffline
	case age > workerFreshFor:
		return domain.WorkerStatusStale
	case activeJobs > 0:
		return domain.WorkerStatusBusy
	default:
		return domain.WorkerStatusOnline
	}
}

func encodeJobTypes(jobTypes []string) string {
	if len(jobTypes) == 0 {
		return ""
	}

	normalized := make([]string, 0, len(jobTypes))
	for _, jobType := range jobTypes {
		jobType = strings.TrimSpace(jobType)
		if jobType == "" {
			continue
		}
		normalized = append(normalized, jobType)
	}
	if len(normalized) == 0 {
		return ""
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	return string(raw)
}

func decodeJobTypes(raw string) []string {
	if raw == "" {
		return nil
	}

	var jobTypes []string
	if err := json.Unmarshal([]byte(raw), &jobTypes); err == nil {
		return jobTypes
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func parseInt(raw string) int {
	if raw == "" {
		return 0
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return val
}

func parseUnix(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || val <= 0 {
		return time.Time{}
	}
	return time.Unix(val, 0)
}

func workerOrder(status domain.WorkerRuntimeStatus) int {
	switch status {
	case domain.WorkerStatusBusy:
		return 0
	case domain.WorkerStatusOnline:
		return 1
	case domain.WorkerStatusStale:
		return 2
	case domain.WorkerStatusOffline:
		return 3
	default:
		return 4
	}
}
