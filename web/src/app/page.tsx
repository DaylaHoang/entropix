import Metric from "@/components/Metric";
import SweepTable from "@/components/SweepTable";
import CostBar from "@/components/CostBar";
import ConfusionMatrix from "@/components/ConfusionMatrix";
import Architecture from "@/components/Architecture";
import EntropyViz from "@/components/EntropyViz";
import StateMachine from "@/components/StateMachine";

function Panel({
  label,
  children,
  className = "",
  badge,
}: {
  label: string;
  children: React.ReactNode;
  className?: string;
  badge?: React.ReactNode;
}) {
  return (
    <div
      className={`rounded-xl border border-(--color-border) bg-(--color-surface) shadow-sm ${className}`}
    >
      <div className="flex items-center justify-between border-b border-(--color-border) px-4 py-3 sm:px-5">
        <span className="font-mono text-[10px] uppercase tracking-[0.15em] text-(--color-text-dark)">
          {label}
        </span>
        {badge}
      </div>
      <div className="p-4 sm:p-5">{children}</div>
    </div>
  );
}

const tradeoffs = [
  {
    title: "CONFIDENT HALLUCINATION",
    color: "var(--color-orange)",
    border: "border-t-[var(--color-orange)]",
    text: "Drafter produces wrong answers with low entropy. 9 false negatives in 518 prompts (1.7%). Mitigated by conservative T, periodic audits, and downstream feedback loops.",
  },
  {
    title: "SPECULATIVE WASTE",
    color: "var(--color-amber)",
    border: "border-t-[var(--color-amber)]",
    text: "Heavyweight fires at 0.8×T but the drafter recovers — wasted parallel call. Occurs on under 10% of escalations. Latency savings on true escalations justify the overhead.",
  },
  {
    title: "CACHE CONSERVATISM",
    color: "var(--color-cyan)",
    border: "border-t-[var(--color-cyan)]",
    text: "Cosine similarity threshold >0.95 is strict by design. Only draft-accepted responses enter the cache; escalated responses are excluded. A stale confident answer is worse than a fresh one.",
  },
  {
    title: "ENTROPY vs CLASSIFICATION",
    color: "var(--color-green)",
    border: "border-t-[var(--color-green)]",
    text: "Prompt classifiers act on input. Entropy acts on the generation itself. This makes routing robust to distribution shift across factual, reasoning, code, and ambiguous prompt categories.",
  },
];

export default function Home() {
  return (
    <div className="min-h-screen">
      <nav className="sticky top-0 z-50 border-b border-(--color-border) bg-(--color-surface) shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
          <div className="flex items-center gap-3">
            <span className="font-mono text-sm font-semibold tracking-tight text-(--color-orange)">
              ENTROPIX
            </span>
            <span className="hidden font-mono text-[10px] text-(--color-text-dark) sm:inline">
              COST-AWARE LLM GATEWAY
            </span>
          </div>
          <div className="flex items-center gap-3 font-mono text-[10px] text-(--color-text-dark) sm:gap-5">
            <span className="hidden items-center gap-1.5 sm:flex">
              <span className="inline-block h-1.5 w-1.5 rounded-full bg-(--color-green)" />
              518 CALIBRATED
            </span>
            <span className="hidden sm:inline">T=2.0</span>
            <a
              href="https://github.com/DaylaHoang/entropix"
              target="_blank"
              rel="noopener noreferrer"
              className="text-(--color-cyan) transition-colors hover:text-(--color-text)"
            >
              SOURCE
            </a>
          </div>
        </div>
      </nav>

      <main className="mx-auto max-w-7xl space-y-4 px-4 py-6 sm:px-6 sm:py-8">

        {/* — Tagline — */}
        <div className="rounded-xl border border-(--color-border) bg-(--color-surface) px-5 py-4 shadow-sm">
          <p className="text-sm leading-relaxed text-(--color-text-dim)">
            Routes{" "}
            <span className="font-semibold text-(--color-green)">94%</span> of
            requests through a cheap model using real-time entropy analysis.
            Escalates only when the drafter&apos;s confidence drops — saving{" "}
            <span className="font-semibold text-(--color-orange)">91.6%</span>{" "}
            in inference costs with{" "}
            <span className="font-semibold text-(--color-cyan)">98.2%</span>{" "}
            accuracy retained.
          </p>
        </div>

        {/* — Key metrics — */}
        <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
          <Metric value="91.6%" label="Cost Reduction"   sublabel="vs all-heavyweight baseline" accent="orange" />
          <Metric value="94.0%" label="Draft Acceptance" sublabel="487 of 518 prompts"           accent="green"  />
          <Metric value="98.2%" label="Draft Accuracy"   sublabel="LLM-as-judge eval"            accent="cyan"   />
          <Metric value="109ms" label="P99 Latency"      sublabel="draft path, 50 req/s"                         />
        </div>

        {/* — System overview: Architecture + Live State Machine — */}
        <div className="grid gap-4 lg:grid-cols-5">
          <Panel label="System Architecture" className="lg:col-span-3">
            <Architecture />
          </Panel>
          <Panel
            label="Routing State Machine"
            className="lg:col-span-2"
          >
            <StateMachine />
            <div className="mt-5 border-t border-(--color-border) pt-4">
              <div className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
                Key Parameters
              </div>
              <div className="mt-3 grid grid-cols-1 gap-x-8 gap-y-1.5 font-mono text-[10px] sm:grid-cols-2">
                {[
                  ["HARD THRESHOLD", "T=2.0 bits",   "var(--color-orange)"],
                  ["SOFT THRESHOLD", "0.8×T=1.6 bits","var(--color-amber)"],
                  ["WINDOW",         "10 tokens",     "var(--color-text-dim)"],
                  ["EARLY EXIT",     "10 tokens",     "var(--color-text-dim)"],
                  ["CACHE COSINE",   ">0.95",          "var(--color-cyan)"],
                  ["TOP LOGPROBS",   "5",             "var(--color-text-dim)"],
                ].map(([k, v, c]) => (
                  <div key={k} className="flex justify-between">
                    <span className="text-(--color-text-dark)">{k}</span>
                    <span style={{ color: c }}>{v}</span>
                  </div>
                ))}
              </div>
            </div>
          </Panel>
        </div>

        {/* — Decision mechanism: Entropy viz + Confusion Matrix — */}
        <div className="grid gap-4 lg:grid-cols-2">
          <Panel label="Entropy Analysis">
            <p className="mb-4 text-xs leading-relaxed text-(--color-text-dark)">
              H(X) = −Σ p(x) log₂ p(x) over top-5 logprobs per token. 10-token
              sliding window. Drag the threshold to see how routing decisions change.
            </p>
            <EntropyViz />
          </Panel>
          <Panel label="Confusion Matrix at T=2.0">
            <ConfusionMatrix />
            <p className="mt-4 text-xs leading-relaxed text-(--color-text-dark)">
              9 false negatives (drafter wrong, low entropy). 29 false positives
              (drafter correct but escalated — safe, just costly). The asymmetry is
              intentional: it is cheaper to over-escalate than to serve a bad draft.
            </p>
          </Panel>
        </div>

        {/* — Cost evidence: CostBar + Model — */}
        <div className="grid gap-4 lg:grid-cols-3">
          <Panel label="Cost Reduction by Threshold" className="lg:col-span-2">
            <CostBar />
          </Panel>
          <Panel label="Cost Model">
            <div className="space-y-4">
              <div>
                <div className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
                  Baseline (all-heavyweight)
                </div>
                <div className="mt-1 font-mono text-xl text-(--color-text)">$1.591</div>
                <div className="text-[10px] text-(--color-text-dark)">518 prompts</div>
              </div>
              <div>
                <div className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
                  Routed at T=2.0
                </div>
                <div className="mt-1 font-mono text-xl text-(--color-green)">$0.133</div>
                <div className="text-[10px] text-(--color-text-dark)">91.6% reduction</div>
              </div>
              <div className="space-y-1.5 border-t border-(--color-border) pt-4">
                {[
                  ["DRAFTER OUT",  "$0.80/1M tok"],
                  ["HEAVY OUT",    "$10.00/1M tok"],
                  ["DRAFTER IN",   "$0.20/1M tok"],
                  ["HEAVY IN",     "$2.50/1M tok"],
                ].map(([k, v]) => (
                  <div key={k} className="flex justify-between font-mono text-[10px]">
                    <span className="text-(--color-text-dark)">{k}</span>
                    <span className="text-(--color-text-dim)">{v}</span>
                  </div>
                ))}
              </div>
              <div className="space-y-1.5 border-t border-(--color-border) pt-4">
                {[
                  ["DRAFTER",    "gpt-4.1-nano"],
                  ["HEAVYWEIGHT","gpt-4.1"],
                  ["EMBEDDINGS", "text-embedding-3-small"],
                ].map(([k, v]) => (
                  <div key={k} className="flex justify-between font-mono text-[10px]">
                    <span className="text-(--color-text-dark)">{k}</span>
                    <span className="text-(--color-text-dim)">{v}</span>
                  </div>
                ))}
              </div>
            </div>
          </Panel>
        </div>

        {/* — Threshold calibration — */}
        <Panel label="Threshold Sweep (n=518)">
          <p className="mb-4 text-xs leading-relaxed text-(--color-text-dark)">
            4 prompt categories: factual, reasoning, code generation, ambiguous/creative.
            Evaluated with LLM-as-judge. Swept T=1.0..2.5 in 0.25 increments. Click
            any row to explore tradeoffs.
          </p>
          <SweepTable />
        </Panel>

        {/* — Known Tradeoffs: 4-column card grid — */}
        <div>
          <div className="mb-3 px-1">
            <span className="font-mono text-[10px] uppercase tracking-[0.15em] text-(--color-text-dark)">
              Known Tradeoffs
            </span>
          </div>
          <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
            {tradeoffs.map((item) => (
              <div
                key={item.title}
                className={`rounded-xl border-t-2 border border-(--color-border) bg-(--color-surface) px-4 py-4 shadow-sm ${item.border}`}
              >
                <div className="font-mono text-[10px] font-semibold uppercase tracking-wider" style={{ color: item.color }}>
                  {item.title}
                </div>
                <p className="mt-2 text-xs leading-relaxed text-(--color-text-dim)">
                  {item.text}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* — Tech stack — */}
        <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-5">
          {[
            ["GATEWAY",    "Go net/http"],
            ["DRAFTER",    "gpt-4.1-nano"],
            ["HEAVYWEIGHT","gpt-4.1"],
            ["CACHE",      "Qdrant + Redis"],
            ["OBS",        "Prometheus + Grafana"],
          ].map(([label, value]) => (
            <div
              key={label}
              className="rounded-xl border border-(--color-border) bg-(--color-surface) px-3 py-2.5 shadow-sm sm:px-4"
            >
              <div className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
                {label}
              </div>
              <div className="mt-0.5 text-xs text-(--color-text-dim)">{value}</div>
            </div>
          ))}
        </div>
      </main>

      <footer className="border-t border-(--color-border)">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-2 px-4 py-4 text-[11px] text-(--color-text-dark) sm:flex-row sm:px-6">
          <span>
            Built by Oliver Vu. All metrics measured on real OpenAI APIs.
            Inspired by{" "}
            <a
              href="https://arxiv.org/abs/2603.00578"
              target="_blank"
              rel="noopener noreferrer"
              className="text-(--color-text-dim) underline decoration-(--color-border-bright) underline-offset-2 transition-colors hover:text-(--color-text)"
            >
              Draft-Thinking (arXiv:2603.00578)
            </a>
            .
          </span>
          <a
            href="https://github.com/DaylaHoang/entropix"
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono text-(--color-cyan) transition-colors hover:text-(--color-text)"
          >
            github.com/DaylaHoang/entropix
          </a>
        </div>
      </footer>
    </div>
  );
}
