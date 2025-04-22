package maprep

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/de4et/your-load/app/internal/worker"
)

type repKey struct {
	CamID     string
	TimeStamp time.Time
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

	r.m[repKey{pr.CamID, pr.Timestamp}] = pr.PeopleAmount
	log.Printf("Wrote result to MapRep: %+v --- %+v", repKey{pr.CamID, pr.Timestamp}, r.m[repKey{pr.CamID, pr.Timestamp}])
	return nil
}
