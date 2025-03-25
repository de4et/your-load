package processor

import "image"

type ImageProcessor interface {
	Process(image.Image) (int, error)
}
