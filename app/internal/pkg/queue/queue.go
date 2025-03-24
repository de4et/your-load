package queue

import (
	"context"
	"time"
)

type ImageQueueElement struct {
	CamID     string
	Timestamp time.Time
	ImageURI  string
}

type ImageQueueAdder interface {
	Add(ctx context.Context, elem ImageQueueElement) error
}

type ImageQueueGetter interface {
	Get(ctx context.Context) (*ImageQueueElement, error)
}
