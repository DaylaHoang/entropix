package sweep

import (
	"github.com/DaylaHoang/entropix/benchmarks/internal/collect"
	"github.com/DaylaHoang/entropix/internal/analysis"
)

type ReplayConfig struct {
	WindowSize     int
	EarlyExitCount int
}

type ReplayResult struct {
	WouldEscalate  bool
	EscalateAtToken int
	TokensConsumed int
}

func Replay(tokens []collect.TokenRecord, threshold float64, cfg ReplayConfig) ReplayResult {
	window := analysis.NewWindow(analysis.WindowConfig{
		Size:           cfg.WindowSize,
		Threshold:      threshold,
		EarlyExitCount: cfg.EarlyExitCount,
	})

	for i, t := range tokens {
		decision := window.Add(t.Entropy)
		if decision == analysis.Escalate {
			return ReplayResult{
				WouldEscalate:   true,
				EscalateAtToken: i,
				TokensConsumed:  i + 1,
			}
		}
	}

	return ReplayResult{
		WouldEscalate:  false,
		TokensConsumed: len(tokens),
	}
}
