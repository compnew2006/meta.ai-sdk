package rest

// job.go implements a tiny in-memory async job store for /video/async +
// /video/jobs/{job_id}. Jobs are stored by id; a background goroutine runs the
// video generation and updates the job state when done.

import (
	"sync"
	"time"

	"github.com/smart-studio/metaai-go/internal/uuid"
)

// jobState values.
const (
	jobQueued    = "queued"
	jobRunning   = "running"
	jobCompleted = "completed"
	jobFailed    = "failed"
)

// job holds the state of one async video generation.
type job struct {
	mu      sync.RWMutex
	id      string
	status  string
	result  *VideoResponse
	err     string
	created time.Time
}

// jobStore is a thread-safe map of job id → *job.
type jobStore struct {
	mu   sync.RWMutex
	jobs map[string]*job
}

func newJobStore() *jobStore {
	return &jobStore{jobs: map[string]*job{}}
}

func (s *jobStore) create() *job {
	j := &job{id: uuid.V4(), status: jobQueued, created: time.Now()}
	s.mu.Lock()
	s.jobs[j.id] = j
	s.mu.Unlock()
	return j
}

func (s *jobStore) get(id string) (*job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

// markRunning transitions a job to the running state.
func (j *job) markRunning() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status = jobRunning
}

// complete transitions a job to completed with the given result.
func (j *job) complete(res *VideoResponse) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status = jobCompleted
	j.result = res
}

// fail transitions a job to failed with the given error.
func (j *job) fail(err string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status = jobFailed
	j.err = err
}

// getInfo returns copy of state, result, and err of a job safely.
func (j *job) getInfo() (string, *VideoResponse, string) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.status, j.result, j.err
}
