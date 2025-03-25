package worker

import (
	"context"
	"sync"

	store "github.com/de4et/your-load/app/internal/worker/imagestore"
	"github.com/de4et/your-load/app/internal/worker/processor"
	"github.com/de4et/your-load/app/internal/worker/queue"
)

type Worker struct {
	imageStore     store.ImageStoreGetter
	imageQueue     queue.ImageQueueGetter
	imageProcessor processor.ImageProcessor

	jobs []*Job
	mu   sync.Mutex
}

func NewWorker(imageStore store.ImageStoreGetter, imageQueue queue.ImageQueueGetter, imageProcessor processor.ImageProcessor) *Worker {
	return &Worker{
		imageStore:     imageStore,
		imageProcessor: imageProcessor,
		imageQueue:     imageQueue,
		jobs:           make([]*Job, 0),
	}
}

func (w *Worker) AddJob(ctx context.Context, job *Job) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.jobs = append(w.jobs, job)
	job.Start(ctx)
}

func (w *Worker) Jobs() int {
	w.updateJobs()
	return len(w.jobs)
}

func (w *Worker) CloseAll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, v := range w.jobs {
		v.Close()
	}
	w.jobs = make([]*Job, 0)
}

func (w *Worker) updateJobs() {
	tmp := make([]*Job, 0, len(w.jobs))
	for _, v := range w.jobs {
		if !v.Closed() {
			tmp = append(tmp, v)
		}
	}
	w.jobs = tmp
}
