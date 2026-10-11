"use client";

import { useCallback, useEffect, useState } from "react";
import { api, km, mins, money, when } from "@/lib/api";
import { downloadFile } from "@/lib/api";
import { Badge, ErrorBox, Spinner, useAction } from "@/components/ui";
import FareLines, { Breakdown } from "@/components/FareLines";
import { Ctx, can } from "./ctx";

export type BookingRow = {
  id: string; reference: string; status: string; tripStatus: string | null; pickupAddress: string; dropoffAddress: string;
  customerName: string; driverName: string | null; scheduledFor: string | null; totalMinor: number; isFinal: boolean; currency: string; createdAt: string;
};

const STATUSES = ["", "scheduled", "searching", "assigned", "completed", "cancelled", "expired"];

export default function Bookings({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [status, setStatus] = useState("");
  const [rows, setRows] = useState<BookingRow[] | null>(null);
  const [open, setOpen] = useState<string | null>(null);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback(() => api<{ bookings: BookingRow[] }>("GET", `${base}/bookings?status=${status}`).then((r) => setRows(r.bookings)).catch(setError), [base, status]);
  useEffect(() => { load(); }, [load]);

  return (
    <div className="split">
      <div className="card">
        <div className="row" style={{ marginBottom: 8 }}>
          <h2 style={{ margin: 0 }}>Bookings</h2>
          {can(ctx.role, "reports.read") && <button className="btn small" onClick={() => downloadFile(`${base}/exports/bookings.csv`, "bookings.csv").catch(() => {})}>Export CSV</button>}
          <select aria-label="Filter by status" className="right" style={{ width: "auto" }} value={status} onChange={(e) => setStatus(e.target.value)}>
            {STATUSES.map((s) => <option key={s} value={s}>{s || "All statuses"}</option>)}
          </select>
        </div>
        <ErrorBox error={error} />
        {!rows ? <Spinner /> : rows.length === 0 ? <div className="empty">No bookings.</div> : <BookingTable rows={rows} onOpen={setOpen} />}
      </div>
      <div>{open ? <BookingDetail key={open} ctx={ctx} id={open} onChange={load} /> : <div className="card empty">Select a booking to see its timeline and fare.</div>}</div>
    </div>
  );
}

export function BookingTable({ rows, onOpen }: { rows: BookingRow[]; onOpen: (id: string) => void }) {
  return (
    <table className="list">
      <tbody>
        {rows.map((b) => (
          <tr key={b.id} className="clickable" onClick={() => onOpen(b.id)}>
            <td data-label="Booking">
              <strong className="mono">{b.reference}</strong> <Badge value={b.tripStatus && b.status === "assigned" ? b.tripStatus : b.status} />
              <div className="small">{b.pickupAddress} → {b.dropoffAddress}</div>
              <div className="muted small">{b.customerName}{b.driverName && ` · driver ${b.driverName}`}{b.scheduledFor && ` · for ${when(b.scheduledFor)}`}</div>
            </td>
            <td data-label="Fare" style={{ textAlign: "right", whiteSpace: "nowrap" }}>
              {money(b.totalMinor, b.currency)}<div className="muted small">{b.isFinal ? "final" : "estimate"}</div>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function BookingDetail({ ctx, id, onChange }: { ctx: Ctx; id: string; onChange: () => void }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [b, setB] = useState<any>(null);
  const act = useAction();
  const load = useCallback(() => api<{ booking: any }>("GET", `${base}/bookings/${id}`).then((r) => setB(r.booking)).catch(act.setError), [base, id, act.setError]);
  useEffect(() => { load(); }, [load]);
  if (!b) return <div className="card"><ErrorBox error={act.error} /><Spinner /></div>;

  const cancellable = ["scheduled", "searching"].includes(b.status) || (b.status === "assigned" && b.trip?.status !== "in_progress");
  return (
    <div className="card stack">
      <div className="row"><h2 className="mono" style={{ margin: 0 }}>{b.reference}</h2><Badge value={b.status} />{b.trip && <Badge value={b.trip.status} />}</div>
      <dl className="kv">
        <dt>Customer</dt><dd>{b.customer?.name} · {b.customer?.phone}</dd>
        <dt>Pickup</dt><dd>{b.pickup.address}</dd>
        <dt>Drop-off</dt><dd>{b.dropoff.address}</dd>
        {b.scheduledFor && <><dt>Scheduled</dt><dd>{when(b.scheduledFor)}</dd></>}
        <dt>Payment</dt><dd>{b.paymentMethod}{b.payments?.length ? ` · collected ${money(b.payments[0].amountMinor, b.payments[0].currency)}` : ""}</dd>
        {b.driver && <><dt>Driver</dt><dd>{b.driver.firstName} · {b.driver.vehiclePlate} ({b.driver.vehicleDescription})</dd></>}
        {b.trip?.distanceM != null && <><dt>Trip</dt><dd>{km(b.trip.distanceM)} · {mins(b.trip.durationS)} · {b.trip.distanceSource === "gps_track" ? "measured by GPS" : "estimated distance"}</dd></>}
        {b.note && <><dt>Note</dt><dd>{b.note}</dd></>}
        {b.cancelReason && <><dt>Cancelled</dt><dd>{b.cancelReason}</dd></>}
      </dl>
      {b.isFinal ? <FareLines b={b.trip.finalBreakdown as Breakdown} title="Final fare" /> : <FareLines b={b.estimate as Breakdown} title="Estimate (final fare is calculated at drop-off)" />}
      <ErrorBox error={act.error} />
      {can(ctx.role, "dispatch.manage") && cancellable && (
        <div><button className="btn danger" disabled={act.busy} onClick={() => {
          const reason = prompt("Why is this booking being cancelled?");
          if (reason) act.run(async () => { await api("POST", `${base}/bookings/${id}/cancel`, { reason }); await load(); onChange(); });
        }}>Cancel booking</button></div>
      )}
      <details>
        <summary>Timeline ({b.events?.length ?? 0} events)</summary>
        <ol className="small" style={{ paddingLeft: 18 }}>
          {b.events?.map((e: any, i: number) => (
            <li key={i}><span className="muted">{when(e.at)}</span> — {e.event.replaceAll("_", " ").replace(".", ": ")}{e.reason && ` (${e.reason})`}</li>
          ))}
        </ol>
      </details>
    </div>
  );
}
