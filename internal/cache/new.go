package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/pkg/llm"
)

func New(
	openaiBaseURL, apiKey, embeddingModel string,
	embeddingDimensions int,
	qdrantURL, qdrantCollection string,
	redisAddr string,
	ttlSeconds int,
	similarityThreshold float64,
	startupTimeout time.Duration,
	rec metrics.Recorder,
) (*VectorStore, error) {
	emb := llm.NewEmbeddingClient(openaiBaseURL, apiKey, embeddingModel, embeddingDimensions, startupTimeout)

	idx := newQdrantClient(qdrantURL, qdrantCollection, startupTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	if err := idx.EnsureCollection(ctx, embeddingDimensions); err != nil {
		return nil, fmt.Errorf("ensuring qdrant collection: %w", err)
	}

	kv, err := newRedisClient(redisAddr, ttlSeconds)
	if err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	return NewVectorStore(emb, idx, kv, similarityThreshold, rec), nil
}
