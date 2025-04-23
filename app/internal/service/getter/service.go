package getter

import (
	"context"
	"fmt"
	"time"

	"github.com/de4et/your-load/app/internal/getter"
	"github.com/de4et/your-load/app/internal/getter/checker"
	"github.com/de4et/your-load/app/internal/getter/downloader"
	store "github.com/de4et/your-load/app/internal/getter/imagestore"
	"github.com/de4et/your-load/app/internal/getter/queue"
)

type GetterService struct {
	imageStore store.ImageStoreAdder
	imageQueue queue.ImageQueueAdder
	// jobRepository
	getter *getter.Getter

	ctx       context.Context
	ctxCancel func()
}

func NewGetterService(
	imageStore store.ImageStoreAdder,
	imageQueue queue.ImageQueueAdder) *GetterService {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &GetterService{
		imageStore: imageStore,
		imageQueue: imageQueue,
		getter: getter.NewGetter(imageStore,
			imageQueue),
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
}

func (s *GetterService) AddJobByDownloader(until time.Time, task getter.Task, downloader downloader.StreamDownloader) error {
	return s.getter.AddJob(s.ctx, getter.NewJob(
		until,
		task,
		downloader,
	))
}

func (s *GetterService) CloseAll() {
	s.getter.CloseAll()
}

func (s *GetterService) Check(url string) (checker.CheckerResponse, error) {
	return s.getter.Check(url)
}

func (s *GetterService) AddJob(until time.Time, task getter.Task) error {
	switch task.Type {
	case checker.ProtocolHLS:
		return s.AddJobByDownloader(until, task, downloader.NewHLSStreamDownloader(task.URL, task.RateInSeconds, 2))
	default:
		return fmt.Errorf("no such downloader for such protocol")
	}
}

func (s *GetterService) Jobs() int {
	return s.getter.Jobs()
}
