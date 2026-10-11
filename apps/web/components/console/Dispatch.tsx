"use client";

import dynamic from "next/dynamic";
import { useCallback, useMemo, useState } from "react";
import { ago, api, newIdempotencyKey } from "@/lib/api";
import { Badge, ErrorBox, usePoll, useAction } from "@/components/ui";
import PlaceInput from "@/components/PlaceInput";
import { Place, Marker, reverseGeocode, MAP_COLORS } from "@/components/MapView";
import { BookingDetail, BookingRow, BookingTable } from "./Bookings";
import { Category, Ctx } from "./ctx";

const MapView = dynamic(() => import("@/components/MapView"), { ssr: false, loading: () => <div className="map tall" /> });

type BoardDriver = { id: string; name: string; availability: string; vehiclePlate: string | null; categoryId: string | null; lat: number | null; lng: number | null;
  locationAt: string | null; locationFresh: boolean; complianceIssues: string[]; hasPendingOffer: boolean };
type Board = { bookings: BookingRow[]; drivers: BoardDriver[]; generatedAt: string };

export default function Dispatch({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [board, setBoard] = useState<Board | null>(null);
  const [open, setOpen] = useState<string | null>(null);
  const [error, setError] = useState<unknown>(null);
  const refresh = useCallback(() => api<Board>("GET", `${base}/dispatch/board`).then((b) => { setBoard(b); setError(null); }).catch(setError), [base]);
  usePoll(refresh, 5000);

  const markers = useMemo<Marker[]>(() => (board?.drivers ?? []).filter((d) => d.lat != null).map((d) => ({
    lat: d.lat!, lng: d.lng!, label: `${d.name} · ${d.availability}${d.locationFresh ? "" : " · location stale"}`,
    color: !d.locationFresh ? MAP_COLORS.stale : d.availability === "available" ? MAP_COLORS.driver : MAP_COLORS.busy,
  })), [board]);

  return (
    <div className="stack">
      <ErrorBox error={error} />
      <div className="split">
        <div className="stack">
          <PhoneBooking ctx={ctx} onBooked={refresh} />
          <div className="card">
            <h2>Open bookings</h2>
            {!board ? null : board.bookings.length === 0 ? <div className="empty">Nothing waiting.</div> : <BookingTable rows={board.bookings} onOpen={setOpen} />}
          </div>
        </div>
        <div className="stack">
          <MapView className="map tall" markers={markers} fitKey={String(markers.length > 0)} />
          <p className="muted small">Teal = available · amber = on a trip · grey = location older than 2 minutes (not live). Updated {board ? ago(board.generatedAt) : "…"}.</p>
          {open && <AssignPanel key={open} base={base} id={open} drivers={board?.drivers ?? []} onDone={refresh} />}
          {open && <BookingDetail key={"d" + open} ctx={ctx} id={open} onChange={refresh} />}
          <div className="card">
            <h2>Drivers online</h2>
            {(board?.drivers.length ?? 0) === 0 ? <div className="empty">No drivers online.</div> : (
              <table className="list"><tbody>
                {board!.drivers.map((d) => (
                  <tr key={d.id}>
                    <td data-label="Driver"><strong>{d.name}</strong> <span className="muted small mono">{d.vehiclePlate}</span></td>
                    <td data-label="Status"><Badge value={d.availability} />{d.hasPendingOffer && <span className="badge warn">offer out</span>}</td>
                    <td data-label="Location">{d.locationFresh ? ago(d.locationAt) : <span className="badge bad">stale · {ago(d.locationAt)}</span>}</td>
                  </tr>
                ))}
              </tbody></table>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function AssignPanel({ base, id, drivers, onDone }: { base: string; id: string; drivers: BoardDriver[]; onDone: () => void }) {
  const [driverId, setDriverId] = useState("");
  const act = useAction();
  const candidates = drivers.filter((d) => d.availability === "available" && d.complianceIssues.length === 0 && !d.hasPendingOffer);
  return (
    <div className="card stack">
      <h3>Offer to a specific driver</h3>
      <p className="muted small">Only for bookings still searching. The driver must accept within 90 seconds; eligibility is checked again on the server.</p>
      <div className="row" style={{ flexWrap: "nowrap" }}>
        <select aria-label="Driver" value={driverId} onChange={(e) => setDriverId(e.target.value)}>
          <option value="">{candidates.length ? "Choose an available driver…" : "No eligible driver available"}</option>
          {candidates.map((d) => <option key={d.id} value={d.id}>{d.name} · {d.vehiclePlate}{d.locationFresh ? "" : " (location stale)"}</option>)}
        </select>
        <button className="btn primary" disabled={!driverId || act.busy}
          onClick={() => act.run(async () => { await api("POST", `${base}/bookings/${id}/assign`, { driverId }); onDone(); })}>Offer</button>
      </div>
      <ErrorBox error={act.error} />
    </div>
  );
}

function PhoneBooking({ ctx, onBooked }: { ctx: Ctx; onBooked: () => void }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [cats, setCats] = useState<Category[] | null>(null);
  const [show, setShow] = useState(false);
  const [customer, setCustomer] = useState({ fullName: "", phone: "" });
  const [pickup, setPickup] = useState<Place | null>(null);
  const [dropoff, setDropoff] = useState<Place | null>(null);
  const [focus, setFocus] = useState<"pickup" | "dropoff">("pickup");
  const [categoryId, setCategoryId] = useState("");
  const [scheduledFor, setScheduledFor] = useState("");
  const [note, setNote] = useState("");
  const [key, setKey] = useState(newIdempotencyKey);
  const act = useAction();

  const openForm = () => {
    setShow(true);
    if (!cats) api<{ categories: Category[] }>("GET", `${base}/vehicle-categories`).then((r) => { setCats(r.categories); setCategoryId(r.categories[0]?.id ?? ""); }).catch(act.setError);
  };
  const pick = async (p: { lat: number; lng: number }) => {
    const place = { ...p, address: await reverseGeocode(p) };
    if (focus === "pickup") { setPickup(place); setFocus("dropoff"); } else setDropoff(place);
  };

  if (!show) return <div className="card row"><div><h2 style={{ margin: 0 }}>Phone booking</h2><p className="muted small" style={{ margin: 0 }}>Book for a caller.</p></div><button className="btn primary right" onClick={openForm}>New booking</button></div>;

  const markers: Marker[] = [
    ...(pickup ? [{ ...pickup, label: "Pickup", color: MAP_COLORS.pickup }] : []),
    ...(dropoff ? [{ ...dropoff, label: "Drop-off", color: MAP_COLORS.dropoff }] : []),
  ];
  return (
    <form className="card stack" onSubmit={(e) => {
      e.preventDefault();
      act.run(async () => {
        if (!pickup || !dropoff) throw new Error("Set both pickup and drop-off.");
        await api("POST", `${base}/bookings`, {
          customer, categoryId, note, pickup, dropoff,
          ...(scheduledFor && { scheduledFor: new Date(scheduledFor).toISOString() }),
        }, { "Idempotency-Key": key });
        setKey(newIdempotencyKey()); setShow(false); setPickup(null); setDropoff(null); setCustomer({ fullName: "", phone: "" }); setNote(""); setScheduledFor("");
        onBooked();
      });
    }}>
      <div className="row"><h2 style={{ margin: 0 }}>Phone booking</h2><button type="button" className="linkbtn right" onClick={() => setShow(false)}>Close</button></div>
      <div className="fields">
        <label className="field">Caller name<input required value={customer.fullName} onChange={(e) => setCustomer({ ...customer, fullName: e.target.value })} /></label>
        <label className="field">Caller phone<input required inputMode="tel" placeholder="077 123 4567" value={customer.phone} onChange={(e) => setCustomer({ ...customer, phone: e.target.value })} /></label>
      </div>
      <PlaceInput label="Pickup" value={pickup} onChange={setPickup} active={focus === "pickup"} onFocus={() => setFocus("pickup")} />
      <PlaceInput label="Drop-off" value={dropoff} onChange={setDropoff} active={focus === "dropoff"} onFocus={() => setFocus("dropoff")} />
      <MapView markers={markers} onPick={pick} fitKey={markers.map((m) => m.lat).join()} />
      <p className="muted small">Tapping the map sets the {focus === "pickup" ? "pickup" : "drop-off"}.</p>
      <div className="fields">
        <label className="field">Vehicle<select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>{cats?.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label>
        <label className="field">When (leave empty for now)<input type="datetime-local" value={scheduledFor} onChange={(e) => setScheduledFor(e.target.value)} /></label>
      </div>
      <label className="field">Note for the driver<input value={note} maxLength={500} onChange={(e) => setNote(e.target.value)} /></label>
      <ErrorBox error={act.error} />
      <button className="btn primary" disabled={act.busy}>Book and dispatch</button>
    </form>
  );
}
