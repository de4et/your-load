package getter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/de4et/your-load/app/internal/getter/checker"
	"github.com/de4et/your-load/app/internal/getter/downloader"
	store "github.com/de4et/your-load/app/internal/getter/imagestore"
	"github.com/de4et/your-load/app/internal/pkg/queue"
)

type Getter struct {
	jobs       []*Job
	checker    *checker.Checker
	imageStore store.ImageStoreAdder
	imageQueue queue.ImageQueueAdder

	mu sync.Mutex
}

func NewGetter(imageStore store.ImageStoreAdder, imageQueue queue.ImageQueueAdder) *Getter {
	return &Getter{
		jobs:       make([]*Job, 0),
		checker:    checker.NewChecker(),
		imageStore: imageStore,
		imageQueue: imageQueue,
	}
}

func (g *Getter) Check(url string) (checker.CheckerResponse, error) {
	return g.checker.CheckURL(url)
}

func (g *Getter) AddJob(job *Job) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.jobs = append(g.jobs, job)
	job.Start()
	go g.runListener(job)
}

func (g *Getter) runListener(job *Job) {
	ctx := context.TODO()
	for {
		resp, err := job.Get()
		if err != nil {
			if errors.Is(err, downloader.ErrClosed) {
				return
			}

			log.Printf("err: %v", err)
			break
		}

		name := fmt.Sprintf("%s_%d", job.Task.CamID, resp.Timestamp.UnixNano())
		uri, err := g.imageStore.Add(ctx, resp.Image, name)
		if err != nil {
			panic(err)
		}
		log.Printf("saved to %s", uri)

		g.imageQueue.Add(context.TODO(), queue.ImageQueueElement{
			Timestamp: resp.Timestamp,
			ImageURI:  uri,
			CamID:     job.Task.CamID,
		})
	}

}

func (g *Getter) Jobs() int {
	g.updateJobs()
	return len(g.jobs)
}

func (g *Getter) CloseAll() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, v := range g.jobs {
		v.Close()
	}
	g.jobs = make([]*Job, 0)
}

func (g *Getter) updateJobs() {
	tmp := make([]*Job, 0, len(g.jobs))
	for _, v := range g.jobs {
		if !v.Closed() {
			tmp = append(tmp, v)
		}
	}
	g.jobs = tmp
}
