package downloader

import (
	"context"
	"image"
)

type StreamDownloader interface {
	Start(context.Context) error
	Get() (image.Image, error)
}
