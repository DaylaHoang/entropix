package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DaylaHoang/entropix/internal/cache"
	"github.com/DaylaHoang/entropix/internal/config"
	"github.com/DaylaHoang/entropix/internal/analysis"
	"github.com/DaylaHoang/entropix/internal/metrics"
	"github.com/DaylaHoang/entropix/internal/router"
	"github.com/DaylaHoang/entropix/internal/parallel"
	"github.com/DaylaHoang/entropix/pkg/llm"
	"github.com/DaylaHoang/entropix/pkg/api"
)

type chatHandler struct {
	drafter           llm.LLMClient
	heavyweight       llm.LLMClient
	router            *router.Router
	executor          *parallel.Executor
	recorder          metrics.Recorder
	entropyCfg        config.EntropyConfig
	specCfg           config.SpeculativeConfig
	cache             cache.Store
	cacheInsertTimeout time.Duration
}

func newChatHandler(drafter, heavyweight llm.LLMClient, rtr *router.Router, exec *parallel.Executor, rec metrics.Recorder, entropyCfg config.EntropyConfig, specCfg config.SpeculativeConfig, cacheStore cache.Store, cacheInsertTimeout time.Duration) *chatHandler {
	return &chatHandler{
		drafter:           drafter,
		heavyweight:       heavyweight,
		router:            rtr,
		executor:          exec,
		recorder:          rec,
		entropyCfg:        entropyCfg,
		specCfg:           specCfg,
		cache:             cacheStore,
		cacheInsertTimeout: cacheInsertTimeout,
	}
}

func (h *chatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		writeBadRequest(w, "method not allowed, use POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req api.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.recorder.RecordError("invalid_request")
		writeBadRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if len(req.Messages) == 0 {
		h.recorder.RecordError("invalid_request")
		writeBadRequest(w, "messages array is required and must not be empty")
		return
	}

	if h.cache != nil {
		resp, err := h.cache.Lookup(r.Context(), req.Messages)
		if err != nil {
			log.Printf("cache lookup error: %v", err)
		} else if resp != nil {
			h.recorder.RecordRoutingDecision("cache_hit")
			h.recorder.RecordRequest(resp.Model, http.StatusOK)
			if req.Stream {
				flusher, ok := w.(http.Flusher)
				if !ok {
					writeInternalError(w, "streaming not supported")
					return
				}
				setRoutingHeaders(w, "cache_hit", start)
				h.replayCachedSSE(w, flusher, resp)
			} else {
				setRoutingHeaders(w, "cache_hit", start)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}
			return
		}
	}

	if req.Stream {
		h.handleStream(w, r, &req, start)
	} else {
		h.handleComplete(w, r, &req, start)
	}
}

func (h *chatHandler) handleComplete(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {
	if h.specCfg.IsEnabled() && h.executor != nil {
		h.handleCompleteSpeculative(w, r, req, start)
		return
	}
	h.handleCompleteSerial(w, r, req, start)
}

func (h *chatHandler) handleCompleteSerial(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {

	drafterReq := h.cloneForDrafter(req)

	ch, err := h.drafter.Stream(r.Context(), drafterReq)
	if err != nil {
		elapsed := time.Since(start)
		h.recorder.RecordUpstreamLatency("drafter", elapsed)
		h.handleUpstreamError(w, err)
		return
	}

	result, err := h.router.Route(r.Context(), ch)
	elapsed := time.Since(start)
	h.recorder.RecordUpstreamLatency("drafter", elapsed)

	if err != nil {
		h.recorder.RecordError("routing_error")
		writeInternalError(w, "routing error")
		log.Printf("routing error: %v", err)
		return
	}

	if result.Decision == analysis.Escalate {
		h.recorder.RecordRoutingDecision("escalate")
		log.Printf("escalating to heavyweight (tokens=%d, avg_entropy=%.3f)",
			result.TokenCount, avgEntropy(result.Entropies))

		hvStart := time.Now()
		resp, err := h.heavyweight.Complete(r.Context(), req)
		h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))

		if err != nil {
			h.handleUpstreamError(w, err)
			return
		}

		h.recorder.RecordRequest(resp.Model, http.StatusOK)
		setRoutingHeaders(w, "escalate", start)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	h.recorder.RecordRoutingDecision("accept")
	resp := assembleResponse(result.DraftChunks)
	h.cacheInsert(req.Messages, resp)
	h.recorder.RecordRequest(resp.Model, http.StatusOK)
	setRoutingHeaders(w, "accept", start)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *chatHandler) handleCompleteSpeculative(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {

	drafterReq := h.cloneForDrafter(req)

	ch, err := h.drafter.Stream(r.Context(), drafterReq)
	if err != nil {
		h.recorder.RecordUpstreamLatency("drafter", time.Since(start))
		h.handleUpstreamError(w, err)
		return
	}

	result, err := h.executor.Execute(r.Context(), ch, req)
	h.recorder.RecordUpstreamLatency("drafter", time.Since(start))

	if err != nil {
		h.recorder.RecordError("routing_error")
		writeInternalError(w, "routing error")
		log.Printf("speculative routing error: %v", err)
		return
	}

	switch result.State {
	case parallel.Escalated:
		h.recorder.RecordRoutingDecision("escalate")
		log.Printf("escalating to heavyweight (tokens=%d, avg_entropy=%.3f)",
			result.TokenCount, avgEntropy(result.Entropies))

		if result.HeavyCh != nil {
			defer result.HeavyCancel()
			var hvChunks []api.StreamChunk
			for sc := range result.HeavyCh {
				if sc.Err != nil {
					h.recorder.RecordError("stream_error")
					log.Printf("speculative heavyweight stream error: %v", sc.Err)
					writeInternalError(w, "heavyweight stream error")
					return
				}
				if sc.Chunk != nil {
					hvChunks = append(hvChunks, *sc.Chunk)
				}
			}
			resp := assembleResponse(hvChunks)
			h.recorder.RecordRequest(resp.Model, http.StatusOK)
			setRoutingHeaders(w, "escalate", start)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		hvStart := time.Now()
		resp, err := h.heavyweight.Complete(r.Context(), req)
		h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))
		if err != nil {
			h.handleUpstreamError(w, err)
			return
		}
		h.recorder.RecordRequest(resp.Model, http.StatusOK)
		setRoutingHeaders(w, "escalate", start)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	default:
		h.recorder.RecordRoutingDecision("accept")
		resp := assembleResponse(result.DraftChunks)
		h.cacheInsert(req.Messages, resp)
		h.recorder.RecordRequest(resp.Model, http.StatusOK)
		setRoutingHeaders(w, "accept", start)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func (h *chatHandler) handleStream(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {
	if h.specCfg.IsEnabled() && h.executor != nil {
		h.handleStreamSpeculative(w, r, req, start)
		return
	}
	h.handleStreamSerial(w, r, req, start)
}

func (h *chatHandler) handleStreamSerial(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {

	drafterReq := h.cloneForDrafter(req)

	ch, err := h.drafter.Stream(r.Context(), drafterReq)
	if err != nil {
		elapsed := time.Since(start)
		h.recorder.RecordUpstreamLatency("drafter", elapsed)
		h.handleUpstreamError(w, err)
		return
	}

	result, err := h.router.Route(r.Context(), ch)
	elapsed := time.Since(start)
	h.recorder.RecordUpstreamLatency("drafter", elapsed)

	if err != nil {
		h.recorder.RecordError("routing_error")
		writeInternalError(w, "routing error")
		log.Printf("routing error: %v", err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeInternalError(w, "streaming not supported")
		return
	}

	if result.Decision == analysis.Escalate {
		h.recorder.RecordRoutingDecision("escalate")
		log.Printf("escalating to heavyweight (tokens=%d, avg_entropy=%.3f)",
			result.TokenCount, avgEntropy(result.Entropies))

		hvStart := time.Now()
		hvCh, err := h.heavyweight.Stream(r.Context(), req)
		if err != nil {
			h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))
			h.handleUpstreamError(w, err)
			return
		}

		setRoutingHeaders(w, "escalate", start)
		h.streamSSE(w, flusher, hvCh, func() {
			h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))
		})
		return
	}

	h.recorder.RecordRoutingDecision("accept")
	resp := assembleResponse(result.DraftChunks)
	h.cacheInsert(req.Messages, resp)
	setRoutingHeaders(w, "accept", start)
	h.replayDraftSSE(w, flusher, result.DraftChunks)
}

func (h *chatHandler) handleStreamSpeculative(w http.ResponseWriter, r *http.Request, req *api.ChatCompletionRequest, start time.Time) {

	drafterReq := h.cloneForDrafter(req)

	ch, err := h.drafter.Stream(r.Context(), drafterReq)
	if err != nil {
		h.recorder.RecordUpstreamLatency("drafter", time.Since(start))
		h.handleUpstreamError(w, err)
		return
	}

	result, err := h.executor.Execute(r.Context(), ch, req)
	h.recorder.RecordUpstreamLatency("drafter", time.Since(start))

	if err != nil {
		h.recorder.RecordError("routing_error")
		writeInternalError(w, "routing error")
		log.Printf("speculative routing error: %v", err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		if result.HeavyCancel != nil {
			result.HeavyCancel()
		}
		writeInternalError(w, "streaming not supported")
		return
	}

	switch result.State {
	case parallel.Escalated:
		h.recorder.RecordRoutingDecision("escalate")
		log.Printf("escalating to heavyweight (tokens=%d, avg_entropy=%.3f)",
			result.TokenCount, avgEntropy(result.Entropies))

		if result.HeavyCh != nil {
			defer result.HeavyCancel()
			setRoutingHeaders(w, "escalate", start)
			h.streamSSE(w, flusher, result.HeavyCh, nil)
			return
		}

		hvStart := time.Now()
		hvCh, err := h.heavyweight.Stream(r.Context(), req)
		if err != nil {
			h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))
			h.handleUpstreamError(w, err)
			return
		}
		setRoutingHeaders(w, "escalate", start)
		h.streamSSE(w, flusher, hvCh, func() {
			h.recorder.RecordUpstreamLatency("heavyweight", time.Since(hvStart))
		})

	default:
		h.recorder.RecordRoutingDecision("accept")
		resp := assembleResponse(result.DraftChunks)
		h.cacheInsert(req.Messages, resp)
		setRoutingHeaders(w, "accept", start)
		h.replayDraftSSE(w, flusher, result.DraftChunks)
	}
}

func (h *chatHandler) streamSSE(w http.ResponseWriter, flusher http.Flusher, ch <-chan llm.StreamChunk, onFirstChunk func()) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	firstChunk := true
	model := "unknown"

	for sc := range ch {
		if sc.Err != nil {
			log.Printf("heavyweight stream error: %v", sc.Err)
			h.recorder.RecordError("stream_error")
			break
		}

		if firstChunk && onFirstChunk != nil {
			onFirstChunk()
			firstChunk = false
		}

		if sc.Chunk != nil {
			model = sc.Chunk.Model
			data, err := json.Marshal(sc.Chunk)
			if err != nil {
				log.Printf("marshal chunk error: %v", err)
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				break
			}
			flusher.Flush()
		}
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
	h.recorder.RecordRequest(model, http.StatusOK)
}

func (h *chatHandler) replayDraftSSE(w http.ResponseWriter, flusher http.Flusher, chunks []api.StreamChunk) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	model := "unknown"
	for _, chunk := range chunks {
		model = chunk.Model
		data, err := json.Marshal(chunk)
		if err != nil {
			log.Printf("marshal chunk error: %v", err)
			continue
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			break
		}
		flusher.Flush()
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
	h.recorder.RecordRequest(model, http.StatusOK)
}

func (h *chatHandler) cloneForDrafter(req *api.ChatCompletionRequest) *api.ChatCompletionRequest {
	clone := *req
	clone.Messages = make([]api.Message, len(req.Messages))
	copy(clone.Messages, req.Messages)
	clone.Stream = true
	logprobs := true
	clone.Logprobs = &logprobs
	topLogprobs := h.entropyCfg.TopLogprobs
	clone.TopLogprobs = &topLogprobs
	return &clone
}

func (h *chatHandler) handleUpstreamError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *llm.UpstreamError:
		h.recorder.RecordError("upstream_error")
		h.recorder.RecordRequest("unknown", e.StatusCode)
		writeUpstreamError(w, e.StatusCode, e.Error())
	case *llm.UpstreamTimeoutError:
		h.recorder.RecordError("upstream_timeout")
		h.recorder.RecordRequest("unknown", http.StatusGatewayTimeout)
		writeError(w, http.StatusGatewayTimeout, e.Error(), "timeout_error")
	default:
		h.recorder.RecordError("internal_error")
		h.recorder.RecordRequest("unknown", http.StatusInternalServerError)
		writeInternalError(w, "internal server error")
		log.Printf("unexpected upstream error: %v", err)
	}
}

func (h *chatHandler) cacheInsert(messages []api.Message, resp *api.ChatCompletionResponse) {
	if h.cache == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.cacheInsertTimeout)
		defer cancel()
		if err := h.cache.Insert(ctx, messages, resp); err != nil {
			log.Printf("cache insert error: %v", err)
		}
	}()
}

func (h *chatHandler) replayCachedSSE(w http.ResponseWriter, flusher http.Flusher, resp *api.ChatCompletionResponse) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for _, choice := range resp.Choices {
		chunk := api.StreamChunk{
			ID:     resp.ID,
			Object: "chat.completion.chunk",
			Model:  resp.Model,
			Choices: []api.Choice{
				{
					Index: choice.Index,
					Delta: &api.Delta{
						Role:    "assistant",
						Content: choice.Message.Content,
					},
				},
			},
		}
		data, err := json.Marshal(chunk)
		if err != nil {
			log.Printf("marshal cached chunk error: %v", err)
			continue
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			break
		}
		flusher.Flush()
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func setRoutingHeaders(w http.ResponseWriter, decision string, start time.Time) {
	w.Header().Set("X-Routing-Decision", decision)
	w.Header().Set("X-Request-Duration-Ms", fmt.Sprintf("%d", time.Since(start).Milliseconds()))
}

func avgEntropy(entropies []analysis.TokenEntropy) float64 {
	if len(entropies) == 0 {
		return 0
	}
	sum := 0.0
	for _, e := range entropies {
		sum += e.Entropy
	}
	return sum / float64(len(entropies))
}
