package queue

import (
	"context"

	"github.com/de4et/your-load/app/internal/getter/queue"
)

type ImageQueueElement = queue.ImageQueueElement

type ImageQueueGetter interface {
	Get(ctx context.Context) (*ImageQueueElement, error)
}
