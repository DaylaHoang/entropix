package cache

import (
	"context"

	"github.com/DaylaHoang/entropix/pkg/api"
)

type Store interface {
	Lookup(ctx context.Context, messages []api.Message) (*api.ChatCompletionResponse, error)
	Insert(ctx context.Context, messages []api.Message, resp *api.ChatCompletionResponse) error
	Evict(ctx context.Context, id string) error
}

type embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type vectorIndex interface {
	Search(ctx context.Context, vector []float64, threshold float64) (*SearchResult, error)
	Upsert(ctx context.Context, id string, vector []float64, payload map[string]string) error
	Delete(ctx context.Context, ids []string) error
	EnsureCollection(ctx context.Context, dimensions int) error
}

type kvStore interface {
	Set(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Del(ctx context.Context, key string) error
	Close() error
}

type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]string
}
