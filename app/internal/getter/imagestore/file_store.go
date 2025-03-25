package store

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

type FileImageStore struct {
}

func NewFileImageStore() *FileImageStore {
	return &FileImageStore{}
}

func (m *FileImageStore) Add(ctx context.Context, img image.Image, name string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("ctx is closed")
	default:
	}

	absName, err := saveToFile(img, name)
	if err != nil {
		return "", err
	}
	return absName, nil
}

func (m *FileImageStore) Get(ctx context.Context, uri string) (image.Image, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx is closed")
	default:
	}

	return readFromFile(uri)
}

func readFromFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return jpeg.Decode(f)
}

func saveToFile(img image.Image, name string) (string, error) {
	fname := "imgs/" + name + ".jpg"
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.Println("saving", fname)

	absPath, err := filepath.Abs(fname)
	if err != nil {
		return "", err
	}

	return absPath, jpeg.Encode(f, img, &jpeg.Options{
		Quality: 100,
	})
}
