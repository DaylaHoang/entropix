package parallel

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

func confidentChunk(token string) api.StreamChunk {
	return api.StreamChunk{
		ID: "chunk-c", Object: "chat.completion.chunk", Model: "gpt-4.1-nano",
		Choices: []api.Choice{
			{
				Index: 0,
				Delta: &api.Delta{Content: token},
				Logprobs: &api.ChoiceLogprobs{
					Content: []api.TokenLogprob{
						{
							Token:   token,
							Logprob: math.Log(0.95),
							TopLogprobs: []api.TopLogprobEntry{
								{Token: token, Logprob: math.Log(0.95)},
								{Token: "alt1", Logprob: math.Log(0.03)},
								{Token: "alt2", Logprob: math.Log(0.02)},
							},
						},
					},
				},
			},
		},
	}
}

func uncertainChunk(token string) api.StreamChunk {
	return api.StreamChunk{
		ID: "chunk-u", Object: "chat.completion.chunk", Model: "gpt-4.1-nano",
		Choices: []api.Choice{
			{
				Index: 0,
				Delta: &api.Delta{Content: token},
				Logprobs: &api.ChoiceLogprobs{
					Content: []api.TokenLogprob{
						{
							Token:   token,
							Logprob: math.Log(0.34),
							TopLogprobs: []api.TopLogprobEntry{
								{Token: token, Logprob: math.Log(0.34)},
								{Token: "alt1", Logprob: math.Log(0.33)},
								{Token: "alt2", Logprob: math.Log(0.33)},
							},
						},
					},
				},
			},
		},
	}
}

// midEntropyChunk generates a chunk with entropy above softThreshold (0.8) but below hardThreshold (1.0).
// Probs 0.80/0.12/0.08 give Shannon entropy ~0.92 bits.
func midEntropyChunk(token string) api.StreamChunk {
	return api.StreamChunk{
		ID: "chunk-m", Object: "chat.completion.chunk", Model: "gpt-4.1-nano",
		Choices: []api.Choice{
			{
				Index: 0,
				Delta: &api.Delta{Content: token},
				Logprobs: &api.ChoiceLogprobs{
					Content: []api.TokenLogprob{
						{
							Token:   token,
							Logprob: math.Log(0.80),
							TopLogprobs: []api.TopLogprobEntry{
								{Token: token, Logprob: math.Log(0.80)},
								{Token: "alt1", Logprob: math.Log(0.12)},
								{Token: "alt2", Logprob: math.Log(0.08)},
							},
						},
					},
				},
			},
		},
	}
}

type mockHeavyweight struct {
	streamChunks []api.StreamChunk
	streamErr    error
	streamCalled bool
}

func (m *mockHeavyweight) Complete(_ context.Context, _ *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	return nil, nil
}

func (m *mockHeavyweight) Stream(_ context.Context, _ *api.ChatCompletionRequest) (<-chan llm.StreamChunk, error) {
	m.streamCalled = true
	if m.streamErr != nil {
		return nil, m.streamErr
	}
	ch := make(chan llm.StreamChunk, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- llm.StreamChunk{Chunk: &m.streamChunks[i]}
	}
	close(ch)
	return ch, nil
}

func defaultCfg() ExecutorConfig {
	return ExecutorConfig{
		WindowCfg: analysis.WindowConfig{
			Size:           5,
			Threshold:      1.0,
			EarlyExitCount: 3,
		},
		SoftThresholdMult: 0.8,
	}
}

func sendChunks(chunks []api.StreamChunk) <-chan llm.StreamChunk {
	ch := make(chan llm.StreamChunk, len(chunks))
	for i := range chunks {
		ch <- llm.StreamChunk{Chunk: &chunks[i]}
	}
	close(ch)
	return ch
}

func testReq() *api.ChatCompletionRequest {
	return &api.ChatCompletionRequest{
		Model:    "auto",
		Messages: []api.Message{{Role: "user", Content: "test"}},
	}
}

func TestExecute_DraftAccepted_NoSoftThreshold(t *testing.T) {
	hw := &mockHeavyweight{}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	chunks := []api.StreamChunk{
		confidentChunk("Hello"),
		confidentChunk(" world"),
	}
	drafterCh := sendChunks(chunks)

	result, err := exec.Execute(context.Background(), drafterCh, testReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State != DraftAccepted {
		t.Errorf("state: got %v, want DraftAccepted", result.State)
	}
	if len(result.DraftChunks) != 2 {
		t.Errorf("chunks: got %d, want 2", len(result.DraftChunks))
	}
	if hw.streamCalled {
		t.Error("heavyweight should not have been called")
	}
	if result.HeavyCh != nil {
		t.Error("HeavyCh should be nil")
	}
}

func TestExecute_DraftAccepted_SoftTriggeredThenRecovered(t *testing.T) {
	hw := &mockHeavyweight{
		streamChunks: []api.StreamChunk{
			{ID: "hw-1", Object: "chat.completion.chunk", Model: "gpt-4.1",
				Choices: []api.Choice{{Index: 0, Delta: &api.Delta{Content: "Heavy"}}}},
		},
	}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	chunks := []api.StreamChunk{
		confidentChunk("Hello"),
		midEntropyChunk("maybe"),
		confidentChunk(" world"),
		confidentChunk(" again"),
		confidentChunk(" more"),
		confidentChunk(" text"),
	}
	drafterCh := sendChunks(chunks)

	result, err := exec.Execute(context.Background(), drafterCh, testReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State != DraftAccepted {
		t.Errorf("state: got %v, want DraftAccepted", result.State)
	}
	if !hw.streamCalled {
		t.Error("heavyweight should have been called speculatively")
	}
	if result.HeavyCh != nil {
		t.Error("HeavyCh should be nil after cancellation")
	}
}

func TestExecute_Escalated_WithSpeculation(t *testing.T) {
	hw := &mockHeavyweight{
		streamChunks: []api.StreamChunk{
			{ID: "hw-1", Object: "chat.completion.chunk", Model: "gpt-4.1",
				Choices: []api.Choice{{Index: 0, Delta: &api.Delta{Content: "Heavy"}}}},
		},
	}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	chunks := []api.StreamChunk{
		midEntropyChunk("hmm"),
		uncertainChunk("well"),
		uncertainChunk("uh"),
	}
	drafterCh := sendChunks(chunks)

	result, err := exec.Execute(context.Background(), drafterCh, testReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State != Escalated {
		t.Errorf("state: got %v, want Escalated", result.State)
	}
	if !hw.streamCalled {
		t.Error("heavyweight should have been called")
	}
	if result.HeavyCh == nil {
		t.Error("HeavyCh should be non-nil for speculative escalation")
	}
	if result.HeavyCancel == nil {
		t.Error("HeavyCancel should be non-nil")
	}
	if result.HeavyCancel != nil {
		result.HeavyCancel()
	}
}

func TestExecute_Escalated_WithoutSpeculation(t *testing.T) {
	hw := &mockHeavyweight{}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	// All uncertain tokens trigger early exit before soft threshold fires
	// (early exit checks individual token entropy > threshold, which happens
	// on the first token before soft threshold logic runs)
	chunks := []api.StreamChunk{
		uncertainChunk("well"),
		uncertainChunk("uh"),
		uncertainChunk("hmm"),
	}
	drafterCh := sendChunks(chunks)

	result, err := exec.Execute(context.Background(), drafterCh, testReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State != Escalated {
		t.Errorf("state: got %v, want Escalated", result.State)
	}
	if result.HeavyCh != nil {
		t.Error("HeavyCh should be nil for early-exit escalation")
	}
}

func TestExecute_HeavyweightStartError(t *testing.T) {
	hw := &mockHeavyweight{
		streamErr: errors.New("connection refused"),
	}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	// midEntropy triggers soft threshold, but heavyweight fails to start.
	// Subsequent confident tokens let drafter complete.
	chunks := []api.StreamChunk{
		midEntropyChunk("hmm"),
		confidentChunk("ok"),
		confidentChunk("fine"),
		confidentChunk("sure"),
		confidentChunk("yes"),
		confidentChunk("done"),
	}
	drafterCh := sendChunks(chunks)

	result, err := exec.Execute(context.Background(), drafterCh, testReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State != DraftAccepted {
		t.Errorf("state: got %v, want DraftAccepted (graceful fallback)", result.State)
	}
}

func TestExecute_DrafterStreamError(t *testing.T) {
	hw := &mockHeavyweight{}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 3)
	c := confidentChunk("ok")
	ch <- llm.StreamChunk{Chunk: &c}
	ch <- llm.StreamChunk{Err: errors.New("stream broke")}
	close(ch)

	_, err := exec.Execute(context.Background(), ch, testReq())
	if err == nil {
		t.Fatal("expected error from drafter stream")
	}
}

func TestExecute_DrafterStreamError_WhileSpeculating(t *testing.T) {
	hw := &mockHeavyweight{
		streamChunks: []api.StreamChunk{
			{ID: "hw-1", Object: "chat.completion.chunk", Model: "gpt-4.1",
				Choices: []api.Choice{{Index: 0, Delta: &api.Delta{Content: "Heavy"}}}},
		},
	}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk, 3)
	c := midEntropyChunk("hmm")
	ch <- llm.StreamChunk{Chunk: &c}
	ch <- llm.StreamChunk{Err: errors.New("stream broke")}
	close(ch)

	_, err := exec.Execute(context.Background(), ch, testReq())
	if err == nil {
		t.Fatal("expected error from drafter stream")
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	hw := &mockHeavyweight{}
	exec := NewExecutor(defaultCfg(), hw, &metrics.NoopRecorder{})

	ch := make(chan llm.StreamChunk)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := exec.Execute(ctx, ch, testReq())
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}
