package store

import (
	"context"
	"image"
)

type ImageStoreAdder interface {
	Add(ctx context.Context, img image.Image, name string) (string, error)
}

type ImageStoreGetter interface {
	Get(ctx context.Context, uri string) (image.Image, error)
}

type ImageStoreDeleter interface {
	Delete(ctx context.Context, uri string) error
}

type ImageStore interface {
	ImageStoreAdder
	ImageStoreGetter
	ImageStoreDeleter
}
