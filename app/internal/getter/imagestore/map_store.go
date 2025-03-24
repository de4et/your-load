package store

import (
	"context"
	"fmt"
	"image"
	"sync"
)

var ErrAlreadyExists = fmt.Errorf("uri is already used")

type MapImageStore struct {
	m  map[string]image.Image
	mu sync.Mutex
}

func NewMapImageStore() *MapImageStore {
	return &MapImageStore{
		m: make(map[string]image.Image),
	}
}

func (m *MapImageStore) Add(ctx context.Context, img image.Image, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("ctx is closed")
	default:
	}

	if _, ok := m.m[name]; ok {
		return "", fmt.Errorf("%v: %s", ErrAlreadyExists, name)
	}

	m.m[name] = img
	return name, nil
}

func (m *MapImageStore) Get(ctx context.Context, uri string) (image.Image, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx is closed")
	default:
	}

	return m.m[uri], nil
}
