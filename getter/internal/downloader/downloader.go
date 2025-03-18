package downloader

import (
	"context"
	"image"
	"time"
)

type downloaderResponse struct {
	Image     image.Image
	Timestamp time.Time
}

type StreamDownloader interface {
	Start(context.Context) error
	Get() (downloaderResponse, error)
	Close()
}
