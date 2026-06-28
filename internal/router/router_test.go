package router

import (
	"context"
	"math"
	"testing"

	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

func makeChunk(token string, topLogprobs []api.TopLogprobEntry) llm.StreamChunk {
	chunk := api.StreamChunk{
		ID:    "test",
		Model: "llama3-8b-8192",
		Choices: []api.Choice{
			{
				Index: 0,
				Delta: &api.Delta{Content: token},
			},
		},
	}

	if topLogprobs != nil {
		mainLogprob := topLogprobs[0].Logprob
		chunk.Choices[0].Logprobs = &api.ChoiceLogprobs{
			Content: []api.TokenLogprob{
				{
					Token:       token,
					Logprob:     mainLogprob,
					TopLogprobs: topLogprobs,
				},
			},
		}
	}

	return llm.StreamChunk{Chunk: &chunk}
}

func confidentLogprobs() []api.TopLogprobEntry {
	return []api.TopLogprobEntry{
		{Token: "the", Logprob: math.Log(0.95)},
		{Token: "a", Logprob: math.Log(0.03)},
		{Token: "an", Logprob: math.Log(0.02)},
	}
}

func uncertainLogprobs() []api.TopLogprobEntry {
	return []api.TopLogprobEntry{
		{Token: "the", Logprob: math.Log(0.34)},
		{Token: "a", Logprob: math.Log(0.33)},
		{Token: "an", Logprob: math.Log(0.33)},
	}
}

func TestRouter_AcceptLowEntropy(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           5,
		Threshold:      1.5,
		EarlyExitCount: 3,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 10)
	for i := 0; i < 10; i++ {
		ch <- makeChunk("token", confidentLogprobs())
	}
	close(ch)

	result, err := r.Route(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != analysis.Continue {
		t.Errorf("expected Continue (accept), got Escalate")
	}
	if len(result.DraftChunks) != 10 {
		t.Errorf("expected 10 chunks, got %d", len(result.DraftChunks))
	}
	if result.TokenCount != 10 {
		t.Errorf("expected 10 tokens, got %d", result.TokenCount)
	}
}

func TestRouter_EscalateHighEntropy(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           5,
		Threshold:      1.0,
		EarlyExitCount: 3,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 10)
	for i := 0; i < 10; i++ {
		ch <- makeChunk("token", uncertainLogprobs())
	}
	close(ch)

	result, err := r.Route(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != analysis.Escalate {
		t.Errorf("expected Escalate, got Continue")
	}
}

func TestRouter_EarlyExit(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           10,
		Threshold:      1.0,
		EarlyExitCount: 5,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 10)
	ch <- makeChunk("hmm", uncertainLogprobs())
	for i := 0; i < 9; i++ {
		ch <- makeChunk("token", confidentLogprobs())
	}
	close(ch)

	result, err := r.Route(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != analysis.Escalate {
		t.Errorf("expected Escalate via early exit, got Continue")
	}
	if result.TokenCount > 1 {
		t.Errorf("expected early exit after 1 token, got %d tokens", result.TokenCount)
	}
}

func TestRouter_NoLogprobs(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           5,
		Threshold:      1.5,
		EarlyExitCount: 3,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 5)
	for i := 0; i < 5; i++ {
		ch <- makeChunk("token", nil) // no logprobs
	}
	close(ch)

	result, err := r.Route(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != analysis.Continue {
		t.Errorf("expected Continue when no logprobs, got Escalate")
	}
	if result.TokenCount != 0 {
		t.Errorf("expected 0 tokens counted (no logprobs), got %d", result.TokenCount)
	}
}

func TestRouter_StreamError(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           5,
		Threshold:      1.5,
		EarlyExitCount: 3,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 2)
	ch <- makeChunk("token", confidentLogprobs())
	ch <- llm.StreamChunk{Err: context.DeadlineExceeded}
	close(ch)

	_, err := r.Route(context.Background(), ch)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRouter_ContextCancellation(t *testing.T) {
	cfg := analysis.WindowConfig{
		Size:           5,
		Threshold:      1.5,
		EarlyExitCount: 3,
	}
	r := NewRouter(cfg, &metrics.NoopRecorder{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan llm.StreamChunk) // unbuffered, will block

	_, err := r.Route(ctx, ch)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
