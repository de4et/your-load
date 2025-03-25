package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/de4et/your-load/app/internal/pkg/queue"
)

type SliceImageQueue struct {
	arr []ImageQueueElement
	mu  sync.Mutex
}

func NewSliceImageQueue() *SliceImageQueue {
	return &SliceImageQueue{
		arr: make([]ImageQueueElement, 0),
	}
}

func (sq *SliceImageQueue) Add(ctx context.Context, elem ImageQueueElement) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx is closed")
	default:
		sq.arr = append(sq.arr, elem)
		return nil
	}
}

func (sq *SliceImageQueue) Get(ctx context.Context) (*ImageQueueElement, error) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx is closed")
	default:
		if len(sq.arr) == 0 {
			return nil, queue.ErrQueueIsEmpty
		}

		el := &sq.arr[0]
		sq.arr = sq.arr[1:]
		return el, nil
	}
}

func (sq *SliceImageQueue) Array() *[]ImageQueueElement {
	return &sq.arr
}
