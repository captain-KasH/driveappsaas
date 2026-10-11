"use client";

import { useCallback, useEffect, useState } from "react";
import { api, money } from "@/lib/api";
import { ErrorBox, Spinner } from "@/components/ui";
import { Ctx, can } from "./ctx";

type Summary = {
  period: { from: string; to: string; timezone: string };
  currency: string; generatedAt: string;
  metrics: { bookings: number; completed: number; cancelled: number; expired: number; completionRate: number | null;
    faresCompletedMinor: number; cashRecordedMinor: number; avgPickupSeconds: number | null; avgRating: number | null };
  definitions: Record<string, string>;
};

export default function Overview({ ctx }: { ctx: Ctx }) {
  const [range, setRange] = useState({ from: "", to: "" });
  const [s, setS] = useState<Summary | null>(null);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback(() => {
    const q = new URLSearchParams(Object.entries(range).filter(([, v]) => v)).toString();
    api<Summary>("GET", `/v1/tenants/${ctx.tenantId}/reports/summary${q ? "?" + q : ""}`).then(setS).catch(setError);
  }, [ctx.tenantId, range]);
  useEffect(() => { if (can(ctx.role, "reports.read")) load(); }, [ctx.role, load]);

  if (!can(ctx.role, "reports.read")) {
    return <div className="card">Welcome. Use the tabs above to work with {ctx.tenant.name}.</div>;
  }
  const m = s?.metrics;
  return (
    <div className="stack">
      <div className="card row">
        <label className="field">From<input type="date" value={range.from} onChange={(e) => setRange({ ...range, from: e.target.value })} /></label>
        <label className="field">To<input type="date" value={range.to} onChange={(e) => setRange({ ...range, to: e.target.value })} /></label>
        <span className="muted small right">{s && `${s.period.from} → ${s.period.to} (${s.period.timezone}) · live data`}</span>
      </div>
      <ErrorBox error={error} />
      {!m ? <Spinner /> : (
        <>
          <div className="grid">
            <div className="card stat"><strong>{m.bookings}</strong><span>bookings</span></div>
            <div className="card stat"><strong>{m.completed}</strong><span>completed · {m.completionRate == null ? "—" : `${Math.round(m.completionRate * 100)}%`}</span></div>
            <div className="card stat"><strong>{m.cancelled + m.expired}</strong><span>cancelled {m.cancelled} · no driver {m.expired}</span></div>
            <div className="card stat"><strong>{money(m.faresCompletedMinor, s!.currency)}</strong><span>fares of completed trips</span></div>
            <div className="card stat"><strong>{money(m.cashRecordedMinor, s!.currency)}</strong><span>cash confirmed by drivers</span></div>
            <div className="card stat"><strong>{m.avgPickupSeconds == null ? "—" : `${Math.round(m.avgPickupSeconds / 60)} min`}</strong><span>avg. assignment → arrival</span></div>
            <div className="card stat"><strong>{m.avgRating == null ? "—" : `${m.avgRating.toFixed(1)} ★`}</strong><span>avg. passenger rating</span></div>
          </div>
          <details className="card small">
            <summary>How these numbers are defined</summary>
            <dl className="kv" style={{ marginTop: 8 }}>
              {Object.entries(s!.definitions).map(([k, v]) => <div key={k} style={{ display: "contents" }}><dt>{k}</dt><dd>{v}</dd></div>)}
            </dl>
          </details>
        </>
      )}
    </div>
  );
}
