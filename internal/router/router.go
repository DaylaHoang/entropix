package router

import (
	"context"
	"fmt"

	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

type RoutingResult struct {
	Decision    analysis.Decision
	DraftChunks []api.StreamChunk
	TokenCount  int
	Entropies   []analysis.TokenEntropy
}

type Router struct {
	windowCfg analysis.WindowConfig
	recorder  metrics.Recorder
}

func NewRouter(cfg analysis.WindowConfig, rec metrics.Recorder) *Router {
	return &Router{
		windowCfg: cfg,
		recorder:  rec,
	}
}

func (r *Router) Route(ctx context.Context, chunks <-chan llm.StreamChunk) (*RoutingResult, error) {
	window := analysis.NewWindow(r.windowCfg)

	result := &RoutingResult{
		Decision: analysis.Continue,
	}

	tokenIndex := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case sc, ok := <-chunks:
			if !ok {
				result.Decision = analysis.Continue
				return result, nil
			}

			if sc.Err != nil {
				return nil, fmt.Errorf("drafter stream error: %w", sc.Err)
			}

			if sc.Chunk == nil {
				continue
			}

			result.DraftChunks = append(result.DraftChunks, *sc.Chunk)

			tokenEntropies := analysis.ExtractTokenEntropy(sc.Chunk, tokenIndex)
			for _, te := range tokenEntropies {
				r.recorder.RecordEntropy(te.Entropy)
				result.Entropies = append(result.Entropies, te)
				result.TokenCount++
				tokenIndex++

				decision := window.Add(te.Entropy)
				if decision == analysis.Escalate {
					result.Decision = analysis.Escalate
					go func() {
						for range chunks {
						}
					}()
					return result, nil
				}
			}

		}
	}
}
