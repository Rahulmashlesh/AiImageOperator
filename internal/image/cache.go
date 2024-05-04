package image

import (
	"context"
	"github.com/go-logr/logr"
)

type Cache interface {
	Get(ctx context.Context, input Input, logger logr.Logger) (Output, bool)
	Remove(ctx context.Context, input Input, logger logr.Logger)
	Insert(ctx context.Context, input Input, logger logr.Logger) (Output, error)
}
