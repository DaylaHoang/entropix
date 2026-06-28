package parallel

import (
	"context"

	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

type State int

const (
	Drafting State = iota
	Speculating
	Escalated
	DraftAccepted
)

func (s State) String() string {
	switch s {
	case Drafting:
		return "drafting"
	case Speculating:
		return "speculating"
	case Escalated:
		return "escalated"
	case DraftAccepted:
		return "draft_accepted"
	default:
		return "unknown"
	}
}

type Result struct {
	State       State
	DraftChunks []api.StreamChunk
	TokenCount  int
	Entropies   []analysis.TokenEntropy
	HeavyCh     <-chan llm.StreamChunk
	HeavyCancel context.CancelFunc
}
