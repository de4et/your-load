package queue

import (
	"context"

	"github.com/de4et/your-load/app/internal/pkg/queue"
)

type ImageQueueElement = queue.ImageQueueElement

type ImageQueueAdder interface {
	Add(ctx context.Context, elem ImageQueueElement) error
}

type ImageQueueGetter interface {
	Get(ctx context.Context) (*ImageQueueElement, error)
}
