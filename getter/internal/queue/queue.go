package queue

import (
	"context"
	"image"
	"time"
)

type ImageQueueElement struct {
	CamID     string
	Timestamp time.Time
	Image     image.Image
}

type ImageQueueAdder interface {
	Add(ctx context.Context, q ImageQueueElement) error
}
