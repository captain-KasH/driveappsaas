"use client";

import { useCallback, useEffect, useState } from "react";
import { api, downloadFile, money, when } from "@/lib/api";
import { ErrorBox, Spinner, useAction } from "@/components/ui";
import { Ctx, can } from "./ctx";

type Customer = { id: string; fullName: string; phone: string; hasAccount: boolean; phoneVerified: boolean; bookings: number; completed: number; cancelled: number; spentMinor: number; lastBookingAt: string | null; createdAt: string };

export default function Customers({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [q, setQ] = useState("");
  const [list, setList] = useState<Customer[] | null>(null);
  const act = useAction();
  const load = useCallback(() => api<{ customers: Customer[] }>("GET", `${base}/customers?q=${encodeURIComponent(q)}`).then((r) => setList(r.customers)).catch(act.setError), [base, q, act.setError]);
  useEffect(() => { const id = setTimeout(load, 250); return () => clearTimeout(id); }, [load]);

  return (
    <div className="card stack">
      <div className="row">
        <h2 style={{ margin: 0 }}>Customers</h2>
        <input className="right" placeholder="Search name or phone" value={q} onChange={(e) => setQ(e.target.value)} style={{ maxWidth: 260 }} />
        {can(ctx.role, "reports.read") && <button className="btn small" disabled={act.busy} onClick={() => act.run(() => downloadFile(`${base}/exports/customers.csv`, "customers.csv"))}>Export CSV</button>}
      </div>
      <p className="muted small">Everyone who has booked with you — from your booking page or by phone. Totals count completed trips only.</p>
      <ErrorBox error={act.error} />
      {!list ? <Spinner /> : list.length === 0 ? <div className="empty">No customers yet.</div> : (
        <table className="list">
          <thead><tr><th>Customer</th><th>Bookings</th><th>Spent</th><th>Last booking</th></tr></thead>
          <tbody>
            {list.map((c) => (
              <tr key={c.id}>
                <td data-label="Customer"><strong>{c.fullName}</strong><div className="muted small mono">{c.phone}{c.hasAccount && " · has app account"}</div></td>
                <td data-label="Bookings">{c.bookings} <span className="muted small">({c.completed} done{c.cancelled ? `, ${c.cancelled} cancelled` : ""})</span></td>
                <td data-label="Spent">{money(c.spentMinor, ctx.tenant.currency)}</td>
                <td data-label="Last booking">{c.lastBookingAt ? when(c.lastBookingAt) : "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
