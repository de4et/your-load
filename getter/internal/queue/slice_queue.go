package queue

import (
	"context"
	"fmt"
)

type SliceImageQueue struct {
	q []ImageQueueElement
}

func NewSliceImageQueue() *SliceImageQueue {
	return &SliceImageQueue{
		q: make([]ImageQueueElement, 0),
	}
}

func (sq *SliceImageQueue) Add(ctx context.Context, elem ImageQueueElement) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx is closed")
	default:
		sq.q = append(sq.q, elem)
		return nil
	}
}

func (sq *SliceImageQueue) Array() []ImageQueueElement {
	return sq.q
}
