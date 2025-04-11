package processor

import (
	"image"
	"math/rand"
)

type StubProcessor struct {
}

func NewStubProcessor() *StubProcessor {
	return &StubProcessor{}
}

func (sp *StubProcessor) Process(image.Image) (int, error) {
	return rand.Int() % 100, nil
}
