package analytics

import (
	"context"
	"time"
)

type WorkerResult struct {
	CamID        string
	TimeStamp    time.Time
	PeopleAmount int
}

type Repository interface {
	Get(ctx context.Context, camID string, timestamp time.Time) (int, error)
	GetByPeriod(ctx context.Context, camID string, start time.Time, end time.Time) ([]WorkerResult, error)
}

type AnalyticsService struct {
	resultRepository Repository

	ctx       context.Context
	ctxCancel func()
}

func NewAnalyticsService(resultRepository Repository) *AnalyticsService {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &AnalyticsService{
		resultRepository: resultRepository,
		ctx:              ctx,
		ctxCancel:        ctxCancel,
	}
}

func (s *AnalyticsService) GetByPeriod(camID string, start time.Time, end time.Time) ([]WorkerResult, error) {
	return s.resultRepository.GetByPeriod(s.ctx, camID, start, end)
}

func (s *AnalyticsService) GetForLast(camID string, duration time.Duration) ([]WorkerResult, error) {
	now := time.Now()
	return s.resultRepository.GetByPeriod(s.ctx, camID, now.Add(-duration), now)
}
