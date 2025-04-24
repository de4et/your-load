package maprep

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/de4et/your-load/app/internal/service/analytics"
	"github.com/de4et/your-load/app/internal/worker"
)

type repKey struct {
	CamID     string
	Timestamp time.Time
}

type MapRepository struct {
	m  map[repKey]int
	mu sync.Mutex
}

func NewMapRepository() *MapRepository {
	return &MapRepository{
		m: make(map[repKey]int),
	}
}

func (r *MapRepository) WriteResult(ctx context.Context, pr worker.ProcResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	r.m[repKey{pr.CamID, pr.Timestamp}] = pr.PeopleAmount
	log.Printf("Wrote result to MapRep: %+v --- %+v", repKey{pr.CamID, pr.Timestamp}, r.m[repKey{pr.CamID, pr.Timestamp}])
	return nil
}

func (r *MapRepository) GetByPeriod(ctx context.Context, camID string, start time.Time, end time.Time) ([]analytics.WorkerResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	res := make([]analytics.WorkerResult, 0)
	for key, value := range r.m {
		if key.CamID == camID && key.Timestamp.After(start) && key.Timestamp.Before(end) {
			res = append(res, analytics.WorkerResult{
				CamID:        camID,
				TimeStamp:    key.Timestamp,
				PeopleAmount: value,
			})
		}
	}

	sort(res)
	return res, nil
}

func (r *MapRepository) Get(ctx context.Context, camID string, timestamp time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return -1, err
	}

	for key, value := range r.m {
		if key.CamID == camID && key.Timestamp.Equal(timestamp) {
			return value, nil
		}
	}

	return -1, fmt.Errorf("no such result for %v in %v", camID, timestamp)
}

func sort(arr []analytics.WorkerResult) {
	slices.SortFunc(arr, func(a analytics.WorkerResult, b analytics.WorkerResult) int {
		return a.TimeStamp.Compare(b.TimeStamp)
	})
}
