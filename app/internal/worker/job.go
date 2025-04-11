package worker

import (
	"context"
	"log"
	"time"

	queuetypes "github.com/de4et/your-load/app/internal/pkg/queue"
	store "github.com/de4et/your-load/app/internal/worker/imagestore"
	"github.com/de4et/your-load/app/internal/worker/processor"
	"github.com/de4et/your-load/app/internal/worker/queue"
)

type ProcResult struct {
	CamID        string
	PeopleAmount int
	Timestamp    time.Time
}

type Repository interface {
	WriteResult(context.Context, ProcResult) error
}

type Job struct {
	imageStore     store.ImageStoreGetter
	imageQueue     queue.ImageQueueGetter
	imageProcessor processor.ImageProcessor
	repository     Repository

	ctx       context.Context
	ctxCancel func()
}

func NewJob(imageStore store.ImageStoreGetter, imageQueue queue.ImageQueueGetter, imageProcessor processor.ImageProcessor, repository Repository) *Job {
	return &Job{
		imageStore:     imageStore,
		imageProcessor: imageProcessor,
		imageQueue:     imageQueue,
		repository:     repository,
	}
}

func (j *Job) Close() {
	j.ctxCancel()
}

func (j *Job) Closed() bool {
	return j.ctx.Err() != nil
}

func (j *Job) Start(ctx context.Context) {
	j.ctx, j.ctxCancel = context.WithCancel(ctx)
	go j.runInner()
}

func (j *Job) runInner() {
	for {
		select {
		case <-j.ctx.Done():
			return
		default:
		}

		el, err := j.imageQueue.Get(j.ctx)
		if err != nil {
			if err == queuetypes.ErrQueueIsEmpty {
				continue
			}
			return
		}

		img, err := j.imageStore.Get(j.ctx, el.ImageURI)
		if err != nil {
			panic(err)
		}

		peopleAmount, err := j.imageProcessor.Process(img)
		if err != nil {
			panic(err)
		}

		res := ProcResult{
			CamID:        el.CamID,
			PeopleAmount: peopleAmount,
			Timestamp:    el.Timestamp,
		}

		err = j.repository.WriteResult(j.ctx, res)
		if err != nil {
			panic(err)
		}

		log.Printf("=!= ID-%s TS-%v PeopleAmount-%d", el.CamID, el.Timestamp, peopleAmount)
	}
}
