package stub

import (
	"context"
	"log"

	"github.com/de4et/your-load/app/internal/worker"
)

type StubRepository struct {
}

func (r *StubRepository) WriteResult(ctx context.Context, pr worker.ProcResult) error {
	log.Printf("Writing result %+v", pr)
	return nil
}
