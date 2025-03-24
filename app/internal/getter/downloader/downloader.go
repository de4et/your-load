package downloader

import (
	"context"
	"image"
	"time"
)

type DownloaderResponse struct {
	Image     image.Image
	Timestamp time.Time
}

type StreamDownloader interface {
	Start(context.Context) error
	Get() (DownloaderResponse, error)
	Close()
}
