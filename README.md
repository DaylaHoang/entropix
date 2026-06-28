# Entropix

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![OpenAI](https://img.shields.io/badge/OpenAI-412991?style=flat-square&logo=openai&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-FF4438?style=flat-square&logo=redis&logoColor=white)
![Qdrant](https://img.shields.io/badge/Qdrant-DC244C?style=flat-square&logo=qdrant&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=flat-square&logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-F46800?style=flat-square&logo=grafana&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)

LLM gateway in Go. Routes each request to a cheap drafter or an expensive heavyweight based on per-token entropy signals — no prompt classifiers, no heuristics. At T=2.0, it cuts inference costs by **91.6%** with **98.2%** accuracy across 518 calibration prompts.

---

## The idea

Most LLM proxies either send everything to the expensive model, or use a classifier on the prompt to decide which model to call. I didn't want a classifier — they break when your prompt distribution shifts, and you don't always know when that's happening.

The thing is, the model already tells you how confident it is. Log-probabilities are right there in every API response. So Entropix watches those signals token by token and escalates to GPT-4.1 only when the drafter gets uncertain. No extra model, no heuristics, just the entropy of the generation itself.

---

## How routing works

Every token comes with log-probabilities for the top 5 candidates. Shannon entropy over those gives a per-token uncertainty score. Entropix tracks a 10-token sliding window and escalates when the mean crosses T.

```
H(X) = -∑ p(x) log₂ p(x)
```

T=2.0 bits came from sweeping T=1.0 to T=2.5 across 518 prompts and finding where accuracy stopped improving. The 10-token window is somewhat arbitrary — I tried 5 and 15, 10 felt right. Short enough to catch problems early, long enough to not react to punctuation noise.

**Speculative execution.** When entropy hits 0.8×T, the heavyweight fires in parallel without waiting for the draft to finish. If the drafter recovers, the heavyweight call gets canceled. If not, it already has a head start. The 0.8 multiplier: at 0.7 you waste too many heavyweight calls on recoveries; at 0.85 the head start is too short to matter.

**Semantic cache.** Before drafting anything, incoming prompts are embedded and compared against cached results via Qdrant. Cosine > 0.95 skips inference entirely. Redis runs alongside for O(1) hash lookup and TTL management — Qdrant alone would work but having a KV store for exact matches is cleaner operationally. Only draft-accepted responses get cached. Escalated responses imply the drafter was shaky on this topic; caching them to serve as confident answers later is exactly the wrong move.

---

## Numbers

518 prompts, 4 categories: factual, multi-step reasoning, code generation, ambiguous/creative. Evaluated with GPT-4.1 as judge.

| | |
|---|---|
| Cost reduction | **91.6%** ($1.591 → $0.133) |
| Draft acceptance | **94%** (487/518) |
| Accuracy at T=2.0 | **98.2%** |
| False negatives | **9** |
| P99 latency, draft path | **109ms** at 50 req/s |
| Gateway overhead | **< 5ms** P99 |

The 9 false negatives are the number I'm not happy with. All of them were multi-step reasoning where the drafter nailed the first few tokens (low entropy) then quietly compounded an error. Entropy can't catch what the model doesn't know it got wrong — no threshold setting fixes this. It needs downstream correctness feedback.

---

## Why Go

Goroutines made the speculative execution pattern nearly trivial to write. Two concurrent LLM streams where you cancel one based on the other's output: `context.WithCancel`, buffered channels for token streaming, `sync.WaitGroup` for clean shutdown. I had a Python prototype first. It worked but the async code was a mess.

No framework in the gateway layer — just `net/http`. The routing logic is a switch statement. It doesn't need to be more than that.

```
cmd/gateway/
internal/
  analysis/      entropy computation, sliding window
  parallel/      speculative executor
  cache/         Qdrant + Redis
  config/        YAML loading, validation
  gateway/       HTTP handler
  metrics/       Prometheus
pkg/
  api/           OpenAI-compatible request/response types
  llm/           LLM and embedding clients
```

`internal/parallel` is where I spent the most time — two streams, shared cancellation, no race conditions. Not a lot of code but getting it right took a few attempts.

---

## Run it

```bash
git clone https://github.com/DaylaHoang/entropix.git
cd entropix
docker compose up -d          # Redis, Qdrant, Prometheus, Grafana

go build -o entropix ./cmd/gateway
export OPENAI_API_KEY=sk-...
./entropix --config config.yaml
```

Grafana at `localhost:3000` (admin/admin) with a pre-built dashboard — entropy distribution, request split, latency percentiles, cache hit rate, speculative overhead.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"What is 2+2?"}]}'
```

**Web demo**

```bash
cd web && npm install && npm run dev
```

Adjust the entropy threshold and see routing decisions update. Has the full calibration dataset if you want to dig into the T tradeoffs.

---

## What doesn't work great

- **Confident hallucination.** 9/518. The drafter can be wrong and certain simultaneously. No threshold setting fixes this.
- **Logprob dependency.** Only works with APIs that expose logprobs. Not every provider does.
- **Cold cache.** Starts empty. Hit rate grows over time — around 12% in testing.
- **Single-node.** Redis and Qdrant run as singletons here. Fine for this, needs work at real scale.

---

**Oliver Vu** — [vvvuminhhoangg@gmail.com](mailto:vvvuminhhoangg@gmail.com) · [github.com/DaylaHoang](https://github.com/DaylaHoang)

MIT
