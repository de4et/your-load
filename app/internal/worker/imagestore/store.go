package store

import (
	"context"
	"image"
)

type ImageStoreGetter interface {
	Get(ctx context.Context, uri string) (image.Image, error)
}
