package llm

import (
	"context"

	"github.com/DaylaHoang/entropix/pkg/api"
)

type StreamChunk struct {
	Chunk *api.StreamChunk
	Err   error
}

type LLMClient interface {
	Complete(ctx context.Context, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error)
	Stream(ctx context.Context, req *api.ChatCompletionRequest) (<-chan StreamChunk, error)
}
