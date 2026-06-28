"use client";

import { useEffect, useRef, useState } from "react";

type MetricProps = {
  value: string;
  label: string;
  sublabel?: string;
  accent?: "orange" | "cyan" | "green" | "default";
};

const borderColors = {
  orange: "border-l-[var(--color-orange)]",
  cyan:   "border-l-[var(--color-cyan)]",
  green:  "border-l-[var(--color-green)]",
  default:"border-l-[var(--color-border-bright)]",
};

const valueColors = {
  orange: "text-[var(--color-orange)]",
  cyan:   "text-[var(--color-cyan)]",
  green:  "text-[var(--color-green)]",
  default:"text-[var(--color-text)]",
};

function parseValue(value: string): { num: number; suffix: string; prefix: string } {
  const match = value.match(/^([^0-9]*)([0-9.]+)(.*)$/);
  if (!match) return { num: 0, suffix: value, prefix: "" };
  return { prefix: match[1], num: parseFloat(match[2]), suffix: match[3] };
}

export default function Metric({ value, label, sublabel, accent = "default" }: MetricProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [displayed, setDisplayed] = useState(0);
  const started = useRef(false);
  const { num, suffix, prefix } = parseValue(value);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !started.current) {
          started.current = true;
          const duration = 900;
          const start = performance.now();
          function tick(now: number) {
            const elapsed = now - start;
            const progress = Math.min(elapsed / duration, 1);
            const eased = 1 - Math.pow(1 - progress, 3);
            setDisplayed(eased * num);
            if (progress < 1) requestAnimationFrame(tick);
          }
          requestAnimationFrame(tick);
        }
      },
      { threshold: 0.3 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [num]);

  const displayStr = Number.isInteger(num)
    ? Math.round(displayed).toString()
    : displayed.toFixed(1);

  return (
    <div
      ref={ref}
      className={`rounded-xl border border-(--color-border) border-l-2 bg-(--color-surface) px-4 py-4 shadow-sm transition-transform duration-200 hover:-translate-y-0.5 sm:px-5 sm:py-5 ${borderColors[accent]}`}
    >
      <div className="font-mono text-[10px] uppercase tracking-widest text-(--color-text-dark)">
        {label}
      </div>
      <div className={`mt-1.5 font-mono text-2xl font-semibold tracking-tight sm:text-3xl ${valueColors[accent]}`}>
        {prefix}{displayStr}{suffix}
      </div>
      {sublabel && (
        <div className="mt-1 text-[11px] text-(--color-text-dark)">{sublabel}</div>
      )}
    </div>
  );
}
