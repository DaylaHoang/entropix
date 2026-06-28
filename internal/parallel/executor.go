package parallel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

type ExecutorConfig struct {
	WindowCfg         analysis.WindowConfig
	SoftThresholdMult float64
}

type Executor struct {
	cfg         ExecutorConfig
	heavyweight llm.LLMClient
	recorder    metrics.Recorder
}

func NewExecutor(cfg ExecutorConfig, hw llm.LLMClient, rec metrics.Recorder) *Executor {
	return &Executor{
		cfg:         cfg,
		heavyweight: hw,
		recorder:    rec,
	}
}

func (e *Executor) Execute(ctx context.Context, drafterCh <-chan llm.StreamChunk, req *api.ChatCompletionRequest) (*Result, error) {
	window := analysis.NewWindow(e.cfg.WindowCfg)
	softThreshold := e.cfg.WindowCfg.Threshold * e.cfg.SoftThresholdMult

	result := &Result{
		State: Drafting,
	}

	var (
		heavyCh     <-chan llm.StreamChunk
		heavyCancel context.CancelFunc
		softTriggered bool
		specStart   time.Time
		tokenIndex  int
	)

	for {
		select {
		case <-ctx.Done():
			if heavyCancel != nil {
				heavyCancel()
				drainChannel(heavyCh)
			}
			return nil, ctx.Err()

		case sc, ok := <-drafterCh:
			if !ok {
				if softTriggered && heavyCancel != nil {
					e.recorder.RecordSpeculativeCancellation()
					heavyCancel()
					drainChannel(heavyCh)
				}
				result.State = DraftAccepted
				return result, nil
			}

			if sc.Err != nil {
				if heavyCancel != nil {
					heavyCancel()
					drainChannel(heavyCh)
				}
				return nil, fmt.Errorf("drafter stream error: %w", sc.Err)
			}

			if sc.Chunk == nil {
				continue
			}

			result.DraftChunks = append(result.DraftChunks, *sc.Chunk)

			tokenEntropies := analysis.ExtractTokenEntropy(sc.Chunk, tokenIndex)
			for _, te := range tokenEntropies {
				e.recorder.RecordEntropy(te.Entropy)
				result.Entropies = append(result.Entropies, te)
				result.TokenCount++
				tokenIndex++

				decision := window.Add(te.Entropy)
				if decision == analysis.Escalate {
					result.State = Escalated
					if softTriggered && heavyCh != nil {
						result.HeavyCh = heavyCh
						result.HeavyCancel = heavyCancel
						e.recorder.RecordSpeculativeLatencySaved(time.Since(specStart))
					}
					go func() {
						for range drafterCh {
						}
					}()
					return result, nil
				}

				if !softTriggered && te.Entropy > softThreshold {
					softTriggered = true
					specStart = time.Now()
					e.recorder.RecordSpeculativeTrigger()

					hvCtx, cancel := context.WithCancel(ctx)
					ch, err := e.heavyweight.Stream(hvCtx, req)
					if err != nil {
						log.Printf("speculative heavyweight start failed: %v", err)
						cancel()
						softTriggered = false
						continue
					}
					heavyCh = ch
					heavyCancel = cancel
					result.State = Speculating
				}
			}
		}
	}
}

func drainChannel(ch <-chan llm.StreamChunk) {
	if ch == nil {
		return
	}
	go func() {
		for range ch {
		}
	}()
}
