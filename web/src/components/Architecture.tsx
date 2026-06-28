export default function Architecture() {
  const f = "var(--font-geist-mono), monospace";

  return (
    <div className="overflow-x-auto">
      <svg
        viewBox="0 0 520 460"
        className="w-full"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          {/* neutral arrow */}
          <marker id="ad" viewBox="0 0 10 8" refX="9" refY="4" markerWidth="6" markerHeight="5" orient="auto">
            <path d="M0 0.5 L8 4 L0 7.5" fill="none" stroke="#9ca3af" strokeWidth="1.2" />
          </marker>
          {/* green arrow */}
          <marker id="ag" viewBox="0 0 10 8" refX="9" refY="4" markerWidth="6" markerHeight="5" orient="auto">
            <path d="M0 0.5 L8 4 L0 7.5" fill="none" stroke="#16a34a" strokeWidth="1.2" />
          </marker>
          {/* orange arrow */}
          <marker id="ao" viewBox="0 0 10 8" refX="9" refY="4" markerWidth="6" markerHeight="5" orient="auto">
            <path d="M0 0.5 L8 4 L0 7.5" fill="none" stroke="#ea580c" strokeWidth="1.2" />
          </marker>
        </defs>

        {/* CLIENT */}
        <rect x={205} y={4} width={100} height={30} rx={2} fill="#ffffff" stroke="#d1d5db" />
        <text x={255} y={23} textAnchor="middle" fill="#111827" fontSize={11} fontFamily={f} fontWeight={500}>
          CLIENT
        </text>

        <line x1={255} y1={34} x2={255} y2={58} stroke="#9ca3af" markerEnd="url(#ad)" />

        {/* GATEWAY */}
        <rect x={180} y={58} width={150} height={36} rx={2} fill="#ffffff" stroke="#d1d5db" />
        <text x={255} y={75} textAnchor="middle" fill="#111827" fontSize={11} fontFamily={f} fontWeight={500}>
          GATEWAY
        </text>
        <text x={255} y={88} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          Go net/http, goroutines
        </text>

        <line x1={255} y1={94} x2={255} y2={120} stroke="#9ca3af" markerEnd="url(#ad)" />

        {/* SEMANTIC CACHE */}
        <rect x={175} y={120} width={160} height={38} rx={2} fill="rgba(37,99,235,0.06)" stroke="#2563eb" strokeOpacity={0.4} />
        <rect x={175} y={120} width={3} height={38} fill="#2563eb" rx={1} />
        <text x={257} y={139} textAnchor="middle" fill="#2563eb" fontSize={11} fontFamily={f} fontWeight={500}>
          SEMANTIC CACHE
        </text>
        <text x={257} y={152} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          {"Qdrant + Redis"}
        </text>

        {/* HIT fork */}
        <path d="M210 158 C210 184, 103 184, 103 206" stroke="#16a34a" strokeOpacity={0.6} fill="none" markerEnd="url(#ag)" />
        <text x={148} y={178} textAnchor="middle" fill="#16a34a" fontSize={9} fontFamily={f} fontWeight={500}>
          HIT
        </text>

        {/* MISS fork */}
        <path d="M300 158 C300 184, 385 184, 385 206" stroke="#9ca3af" strokeOpacity={0.7} fill="none" markerEnd="url(#ad)" />
        <text x={348} y={178} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          MISS
        </text>

        {/* RETURN CACHED */}
        <rect x={28} y={206} width={150} height={30} rx={2} fill="rgba(22,163,74,0.07)" stroke="#16a34a" strokeOpacity={0.35} />
        <rect x={28} y={206} width={3} height={30} fill="#16a34a" rx={1} />
        <text x={105} y={225} textAnchor="middle" fill="#16a34a" fontSize={11} fontFamily={f} fontWeight={500}>
          RETURN CACHED
        </text>

        {/* DRAFTER */}
        <rect x={314} y={206} width={142} height={36} rx={2} fill="rgba(22,163,74,0.06)" stroke="#16a34a" strokeOpacity={0.4} />
        <rect x={314} y={206} width={3} height={36} fill="#16a34a" rx={1} />
        <text x={387} y={223} textAnchor="middle" fill="#16a34a" fontSize={11} fontFamily={f} fontWeight={500}>
          DRAFTER
        </text>
        <text x={387} y={236} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          gpt-4.1-nano, logprobs
        </text>

        <line x1={385} y1={242} x2={385} y2={270} stroke="#9ca3af" markerEnd="url(#ad)" />

        {/* ENTROPY ENGINE */}
        <rect x={296} y={270} width={178} height={38} rx={2} fill="#ffffff" stroke="#374151" strokeOpacity={0.5} />
        <rect x={296} y={270} width={3} height={38} fill="#374151" rx={1} />
        <text x={385} y={289} textAnchor="middle" fill="#111827" fontSize={11} fontFamily={f} fontWeight={500}>
          ENTROPY ENGINE
        </text>
        <text x={385} y={302} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          {"H(X) = -Σ p log₂ p, 10-tok"}
        </text>

        {/* H < T fork */}
        <path d="M340 308 C340 334, 276 334, 276 356" stroke="#16a34a" strokeOpacity={0.6} fill="none" markerEnd="url(#ag)" />
        <text x={300} y={328} textAnchor="middle" fill="#16a34a" fontSize={9} fontFamily={f} fontWeight={500}>
          {"H < T"}
        </text>

        {/* H >= T fork */}
        <path d="M430 308 C430 334, 456 334, 456 356" stroke="#ea580c" strokeOpacity={0.6} fill="none" markerEnd="url(#ao)" />
        <text x={448} y={328} textAnchor="middle" fill="#ea580c" fontSize={9} fontFamily={f} fontWeight={500}>
          {"H ≥ T"}
        </text>

        {/* ACCEPT */}
        <rect x={216} y={356} width={120} height={36} rx={2} fill="rgba(22,163,74,0.07)" stroke="#16a34a" strokeOpacity={0.35} />
        <rect x={216} y={356} width={3} height={36} fill="#16a34a" rx={1} />
        <text x={278} y={373} textAnchor="middle" fill="#16a34a" fontSize={11} fontFamily={f} fontWeight={500}>
          ACCEPT
        </text>
        <text x={278} y={386} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          serve draft, cache
        </text>

        {/* ESCALATE */}
        <rect x={408} y={356} width={95} height={30} rx={2} fill="rgba(234,88,12,0.07)" stroke="#ea580c" strokeOpacity={0.35} />
        <rect x={408} y={356} width={3} height={30} fill="#ea580c" rx={1} />
        <text x={457} y={375} textAnchor="middle" fill="#ea580c" fontSize={11} fontFamily={f} fontWeight={500}>
          ESCALATE
        </text>

        <line x1={456} y1={386} x2={456} y2={408} stroke="#ea580c" strokeOpacity={0.6} markerEnd="url(#ao)" />

        {/* HEAVYWEIGHT */}
        <rect x={396} y={408} width={118} height={36} rx={2} fill="rgba(220,38,38,0.06)" stroke="#dc2626" strokeOpacity={0.35} />
        <rect x={396} y={408} width={3} height={36} fill="#dc2626" rx={1} />
        <text x={457} y={425} textAnchor="middle" fill="#dc2626" fontSize={11} fontFamily={f} fontWeight={500}>
          HEAVYWEIGHT
        </text>
        <text x={457} y={438} textAnchor="middle" fill="#6b7280" fontSize={9} fontFamily={f}>
          gpt-4.1
        </text>

        <text x={455} y={456} textAnchor="middle" fill="#d97706" fontSize={8} fontFamily={f}>
          speculate at 0.8*T
        </text>
      </svg>
    </div>
  );
}
