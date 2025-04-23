package getter

import "github.com/de4et/your-load/app/internal/getter/checker"

type Task struct {
	CamID         string
	URL           string
	RateInSeconds float64
	Type          checker.ProtocolType
}
