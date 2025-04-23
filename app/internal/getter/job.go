package getter

import (
	"context"
	"time"

	"github.com/de4et/your-load/app/internal/getter/downloader"
)

type Job struct {
	ID    string // cam id
	Until time.Time
	Task  Task

	downloader downloader.StreamDownloader
	ctx        context.Context
	ctxCancel  func()
}

func NewJob(until time.Time, task Task, downloader downloader.StreamDownloader) *Job {
	return &Job{
		Until:      until,
		Task:       task,
		downloader: downloader,
	}
}

func (j *Job) Close() {
	j.ctxCancel()
	j.downloader.Close()
}

func (j *Job) Closed() bool {
	return j.ctx.Err() != nil
}

func (j *Job) Start(ctx context.Context) error {
	j.ctx, j.ctxCancel = context.WithDeadline(ctx, j.Until)
	return j.downloader.Start(j.ctx)
}

func (j *Job) Get() (downloader.DownloaderResponse, error) {
	return j.downloader.Get()
}
