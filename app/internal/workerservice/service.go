package workerservice

import (
	"context"

	"github.com/de4et/your-load/app/internal/worker"
	store "github.com/de4et/your-load/app/internal/worker/imagestore"
	"github.com/de4et/your-load/app/internal/worker/processor"
	"github.com/de4et/your-load/app/internal/worker/queue"
)

type WorkerService struct {
	imageStore       store.ImageStoreGetter
	imageQueue       queue.ImageQueueGetter
	imageProcessor   processor.ImageProcessor
	resultRepository worker.Repository
	worker           *worker.Worker

	ctx       context.Context
	ctxCancel func()
}

func NewWorkerService(
	imageStore store.ImageStoreGetter,
	imageQueue queue.ImageQueueGetter,
	imageProcessor processor.ImageProcessor,
	resultRepository worker.Repository) *WorkerService {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &WorkerService{
		imageStore:     imageStore,
		imageProcessor: imageProcessor,
		imageQueue:     imageQueue,
		worker: worker.NewWorker(imageStore,
			imageQueue,
			imageProcessor),
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
}

func (s *WorkerService) AddJob() {
	s.worker.AddJob(s.ctx, worker.NewJob(
		s.imageStore,
		s.imageQueue,
		s.imageProcessor,
		s.resultRepository,
	))
}

func (s *WorkerService) Insert() {
}
