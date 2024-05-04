package image

import (
	"context"
	"github.com/go-logr/logr"
)

type Input struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
	Seed   string `json:"seed,omitempty"`
}

type Output struct {
	Data []byte
	Seed string
}

type Generator interface {
	Generate(ctx context.Context, input Input, logger logr.Logger) (Output, error)
}
