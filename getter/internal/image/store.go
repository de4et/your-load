package image

import (
	"context"
	"image"
)

type StoreAdder interface {
	Add(ctx context.Context, img image.Image, name string) (string, error)
}

type StoreGetter interface {
	Get(ctx context.Context, uri string) (image.Image, error)
}

type StoreDeleter interface {
	Delete(ctx context.Context, uri string) error
}

type Store interface {
	StoreAdder
	StoreGetter
	StoreDeleter
}
