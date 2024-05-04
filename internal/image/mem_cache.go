package image

import (
	"context"
	"github.com/go-logr/logr"
)

type MemCache struct {
	Generator Generator
	mem       map[Input]Output
}

func NewMemCache(generator Generator) *MemCache {
	return &MemCache{
		Generator: generator,
		mem:       make(map[Input]Output),
	}
}

func (m *MemCache) Insert(ctx context.Context, input Input, logger logr.Logger) (Output, error) {
	logger.Info("Trying to insert to cache")
	logger.Info("Trying to Generate Image")
	output, err := m.Generator.Generate(ctx, input, logger)
	if err != nil {
		return Output{}, err
	}

	input.Seed = output.Seed
	m.mem[input] = output
	logger.Info("Inserted into cache")
	return output, nil
}

func (m *MemCache) Get(ctx context.Context, input Input, logger logr.Logger) (Output, bool) {
	logger.Info("Trying to get from cache")
	o, ok := m.mem[input]
	return o, ok
}

func (m *MemCache) Remove(ctx context.Context, input Input, logger logr.Logger) {
	logger.Info("Trying to remove from cache")
	delete(m.mem, input)
}
