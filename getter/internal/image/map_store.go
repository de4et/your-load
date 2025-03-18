package image

import (
	"context"
	"fmt"
	"image"
	"sync"
)

var ErrAlreadyExists = fmt.Errorf("uri is already used")

type MapStore struct {
	nextInt int

	m  map[string]image.Image
	mu sync.Mutex
}

func NewMapStore() *MapStore {
	return &MapStore{
		m: make(map[string]image.Image),
	}
}

func (m *MapStore) Add(ctx context.Context, img image.Image, name string) (string, error) {
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

func (m *MapStore) Get(ctx context.Context, uri string) (image.Image, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx is closed")
	default:
	}

	return m.m[uri], nil
}
