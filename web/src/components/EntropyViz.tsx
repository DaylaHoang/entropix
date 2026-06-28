"use client";

import { useState } from "react";

const examples = [
  { label: "ARITHMETIC", prompt: "What is 347 + 892?",  mean: 0.028, max: 0.303 },
  { label: "FACTUAL",    prompt: "What is photosynthesis?", mean: 0.21,  max: 1.835 },
  { label: "AMBIGUOUS",  prompt: "Define ubiquitous",    mean: 0.359, max: 2.198 },
];

const MAX = 3.0;

export default function EntropyViz() {
  const [threshold, setThreshold] = useState(2.0);

  return (
    <div className="space-y-5">
      <div className="flex items-center gap-3">
        <span className="shrink-0 font-mono text-xs text-(--color-text-dark)">T =</span>
        <span className="w-10 shrink-0 font-mono text-sm font-semibold text-(--color-orange)">
          {threshold.toFixed(1)}
        </span>
        <input
          type="range"
          min={0.5}
          max={3.0}
          step={0.1}
          value={threshold}
          onChange={(e) => setThreshold(parseFloat(e.target.value))}
          className="flex-1 accent-[var(--color-orange)]"
        />
        <span className="shrink-0 font-mono text-[10px] text-(--color-text-dark)">3.0</span>
      </div>

      <div className="space-y-4">
        {examples.map((ex) => {
          const decision = ex.max >= threshold ? "ESCALATE" : "ACCEPT";
          return (
            <div key={ex.prompt}>
              <div className="mb-1.5 flex items-baseline justify-between gap-2">
                <div className="min-w-0">
                  <span className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
                    {ex.label}
                  </span>
                  <span className="ml-2 text-xs text-(--color-text-dim) max-sm:hidden">
                    {ex.prompt}
                  </span>
                </div>
                <span
                  className={`shrink-0 rounded px-1.5 py-0.5 font-mono text-[10px] font-medium transition-colors ${
                    decision === "ACCEPT"
                      ? "bg-(--color-green-dim) text-(--color-green)"
                      : "bg-(--color-orange-dim) text-(--color-orange)"
                  }`}
                >
                  {decision}
                </span>
              </div>
              <div className="relative h-7 rounded bg-(--color-surface-2)">
                <div
                  className="absolute left-0 top-0 h-full rounded-l bg-(--color-cyan-dim) transition-all"
                  style={{ width: `${(ex.mean / MAX) * 100}%` }}
                />
                <div
                  className={`absolute top-0 h-full w-0.5 transition-colors ${
                    ex.max >= threshold ? "bg-(--color-orange)" : "bg-(--color-green)"
                  }`}
                  style={{ left: `${(ex.max / MAX) * 100}%` }}
                />
                <div
                  className="absolute top-0 h-full w-px bg-(--color-red) opacity-40 transition-all"
                  style={{ left: `${(threshold / MAX) * 100}%` }}
                />
                <div className="absolute right-2 top-1/2 flex -translate-y-1/2 gap-3 font-mono text-[10px]">
                  <span className="text-(--color-text-dark)">avg {ex.mean.toFixed(3)}</span>
                  <span className="text-(--color-text-dim)">peak {ex.max.toFixed(3)}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-(--color-border) pt-3 font-mono text-[10px] text-(--color-text-dark)">
        <span className="flex items-center gap-1.5">
          <span className="inline-block h-2 w-px bg-(--color-red) opacity-40" /> T threshold
        </span>
        <span className="flex items-center gap-1.5">
          <span className="inline-block h-2 w-2 rounded-sm bg-(--color-cyan-dim)" /> mean entropy
        </span>
        <span className="flex items-center gap-1.5">
          <span className="inline-block h-2 w-0.5 bg-(--color-green)" /> peak (under T)
        </span>
        <span className="flex items-center gap-1.5">
          <span className="inline-block h-2 w-0.5 bg-(--color-orange)" /> peak (over T)
        </span>
      </div>
      <p className="text-[11px] text-(--color-text-dark)">
        Drag the slider to see how threshold T controls which requests get accepted or escalated.
      </p>
    </div>
  );
}
