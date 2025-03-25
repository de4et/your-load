package queue

import (
	"fmt"
	"time"
)

type ImageQueueElement struct {
	CamID     string
	Timestamp time.Time
	ImageURI  string
}

var ErrQueueIsEmpty = fmt.Errorf("queue is empty")
