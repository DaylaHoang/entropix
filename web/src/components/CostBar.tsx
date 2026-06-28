"use client";

import { useState } from "react";

const data = [
  { t: 1.0,  cost: 8.2,  acc: 100.0 },
  { t: 1.25, cost: 31.0, acc: 99.6  },
  { t: 1.5,  cost: 56.2, acc: 98.6  },
  { t: 1.75, cost: 81.2, acc: 98.4  },
  { t: 2.0,  cost: 91.6, acc: 98.2  },
  { t: 2.25, cost: 99.0, acc: 97.9  },
  { t: 2.5,  cost: 99.2, acc: 97.9  },
];

export default function CostBar() {
  const [selected, setSelected] = useState(2.0);
  const [hovered, setHovered]   = useState<number | null>(null);

  return (
    <div className="space-y-1.5">
      <div className="mb-3 flex font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
        <span className="w-14 shrink-0">T</span>
        <span className="flex-1">Cost Reduction</span>
        <span className="w-14 text-right">ACC</span>
      </div>
      {data.map((d) => {
        const sel  = d.t === selected;
        const hov  = d.t === hovered;
        const base = d.t === 2.0;
        const diff = d.cost - 91.6;

        return (
          <div
            key={d.t}
            className={`group relative flex cursor-pointer items-center rounded transition-colors ${hov ? "bg-(--color-surface-2)" : ""}`}
            onClick={() => setSelected(d.t)}
            onMouseEnter={() => setHovered(d.t)}
            onMouseLeave={() => setHovered(null)}
          >
            <span
              className={`w-14 shrink-0 font-mono text-xs transition-colors ${
                sel ? "font-semibold text-(--color-orange)" : "text-(--color-text-dark)"
              }`}
            >
              {d.t.toFixed(2)}
            </span>
            <div className="relative h-6 flex-1 rounded bg-(--color-surface-2)">
              <div
                className={`h-full rounded transition-all duration-300 ${
                  sel
                    ? "bg-(--color-orange)"
                    : hov
                    ? "bg-(--color-border-bright)"
                    : "bg-(--color-border)"
                }`}
                style={{ width: `${d.cost}%` }}
              />
              {sel && (
                <span className="absolute left-2 top-1/2 -translate-y-1/2 font-mono text-[10px] font-semibold text-white">
                  {d.cost}%
                </span>
              )}
            </div>
            <span
              className={`w-14 text-right font-mono text-xs ${
                d.acc >= 98 ? "text-(--color-green)" : "text-(--color-text-dark)"
              }`}
            >
              {d.acc}%
            </span>

            {hov && (
              <div className="absolute right-16 top-1/2 z-10 -translate-y-1/2 rounded-lg border border-(--color-border) bg-(--color-surface) px-3 py-2 shadow-lg">
                <div className="space-y-0.5 font-mono text-[10px]">
                  <div className="text-(--color-text-dark)">
                    Cost saved:{" "}
                    <span className="text-(--color-orange)">{d.cost}%</span>
                  </div>
                  <div className="text-(--color-text-dark)">
                    Accuracy:{" "}
                    <span className={d.acc >= 98 ? "text-(--color-green)" : "text-(--color-text)"}>
                      {d.acc}%
                    </span>
                  </div>
                  {!base && (
                    <div className="text-(--color-text-dark)">
                      vs T=2.0:{" "}
                      <span className={diff > 0 ? "text-(--color-green)" : "text-(--color-red)"}>
                        {diff > 0 ? "+" : ""}{diff.toFixed(1)}%
                      </span>
                    </div>
                  )}
                  {base && (
                    <div className="text-(--color-cyan)">← calibrated</div>
                  )}
                </div>
              </div>
            )}
          </div>
        );
      })}
      <p className="pt-1 text-[11px] text-(--color-text-dark)">
        Click a row to select threshold. Hover for details.
      </p>
    </div>
  );
}
