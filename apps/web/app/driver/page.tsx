"use client";

import Link from "next/link";
import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { ago, api, km, money } from "@/lib/api";
import { Badge, ErrorBox, SignOut, Spinner, TopBar, useAction, usePoll } from "@/components/ui";
import FareLines from "@/components/FareLines";
import { ISSUE_TEXT } from "@/components/console/ctx";

type DriverMe = {
  driver: { id: string; fullName: string; status: string; availability: string; complianceIssues: string[]; locationAt: string | null; locationFresh: boolean };
  documents: { id: string; docType: string; number: string; expiresOn: string | null; status: string; reviewNote: string | null }[];
  vehicle: { plate: string; make: string; model: string; colour: string } | null;
};
type Offer = { id: string; bookingId: string; expiresAt: string; distanceToPickupM: number | null; pickupAddress: string; dropoffAddress: string;
  estimatedTotalMinor: number; currency: string; pickupLat: number; pickupLng: number; scheduledFor: string | null; customerFirstName: string };

const NEXT: Record<string, { action: string; label: string }[]> = {
  driver_assigned: [{ action: "en-route", label: "I'm on my way" }, { action: "arrived", label: "I've arrived" }],
  en_route: [{ action: "arrived", label: "I've arrived" }],
  arrived: [{ action: "start", label: "Passenger on board — start trip" }],
  in_progress: [{ action: "complete", label: "Complete trip" }],
};
const DOC_TYPES = ["driving_licence", "national_id", "vehicle_insurance", "revenue_licence", "police_report"];
const directions = (lat: number, lng: number) => `https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`;

function DriverApp() {
  const t = useSearchParams().get("t") ?? "";
  const base = `/v1/tenants/${t}`;
  const [me, setMe] = useState<DriverMe | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [trip, setTrip] = useState<any>(null);
  const [finished, setFinished] = useState<any>(null); // completed booking awaiting cash confirmation
  const [geo, setGeo] = useState<"off" | "waiting" | "ok" | "denied" | "unsupported">("off");
  const act = useAction();

  const refresh = useCallback(async () => {
    try {
      const [m, w] = await Promise.all([api<DriverMe>("GET", `${base}/driver/me`), api<{ offer: Offer | null; activeTrip: any }>("GET", `${base}/driver/work`)]);
      setMe(m); setOffer(w.offer); setTrip(w.activeTrip);
    } catch (e) { act.setError(e); }
  }, [base, act.setError]);
  usePoll(refresh, 4000);

  // Share location while online or on a trip; throttled to one update every 10 s.
  const sharing = me && (me.driver.availability !== "offline" || trip);
  const last = useRef(0);
  useEffect(() => {
    if (!sharing) { setGeo("off"); return; }
    if (!("geolocation" in navigator)) { setGeo("unsupported"); return; }
    setGeo("waiting");
    const id = navigator.geolocation.watchPosition(
      (p) => {
        setGeo("ok");
        if (Date.now() - last.current < 10_000) return;
        last.current = Date.now();
        api("POST", `${base}/driver/location`, { lat: p.coords.latitude, lng: p.coords.longitude, recordedAt: new Date(p.timestamp).toISOString() }).catch(() => {});
      },
      (err) => setGeo(err.code === err.PERMISSION_DENIED ? "denied" : "waiting"),
      { enableHighAccuracy: true, maximumAge: 5000, timeout: 20000 },
    );
    return () => navigator.geolocation.clearWatch(id);
  }, [sharing, base]);

  if (!t) return <div className="alert error">No business selected. <Link href="/home/">Choose one</Link>.</div>;
  if (!me) return act.error ? <ErrorBox error={act.error} /> : <Spinner />;

  const d = me.driver;
  const online = d.availability !== "offline";
  const go = (fn: () => Promise<unknown>) => act.run(async () => { await fn(); await refresh(); });

  return (
    <div className="stack">
      <div className="card stack">
        <div className="row">
          <h1 style={{ margin: 0 }}>{d.fullName.split(" ")[0]}</h1>
          <Badge value={d.availability} />
          {me.vehicle && <span className="muted small mono">{me.vehicle.plate}</span>}
        </div>
        {d.complianceIssues.length > 0 ? (
          <div className="alert warn">You can't go online yet: {d.complianceIssues.map((i) => ISSUE_TEXT[i] ?? i).join(", ")}. Add your documents below; your operator reviews them.</div>
        ) : !trip && (
          <button className={`btn block ${online ? "" : "primary"}`} disabled={act.busy}
            onClick={() => go(() => api("POST", `${base}/driver/availability`, { available: !online }))}>
            {online ? "Go offline" : "Go online"}
          </button>
        )}
        {sharing && (
          <div className="small">
            {geo === "ok" && <span className="badge ok">sharing location</span>}
            {geo === "waiting" && <span className="badge warn">waiting for GPS…</span>}
            {geo === "denied" && <div className="alert error">Location is blocked. Allow location for this site, or you won't receive nearby offers.</div>}
            {geo === "unsupported" && <div className="alert error">This browser can't share location.</div>}
            <span className="muted"> · last update {ago(d.locationAt)}</span>
          </div>
        )}
        <ErrorBox error={act.error} />
      </div>

      {offer && !trip && <OfferCard offer={offer} busy={act.busy}
        onAccept={() => go(() => api("POST", `${base}/driver/offers/${offer.id}/accept`))}
        onReject={() => go(() => api("POST", `${base}/driver/offers/${offer.id}/reject`))} />}

      {trip && (
        <div className="card stack">
          <div className="row"><h2 style={{ margin: 0 }}>Current trip</h2><Badge value={trip.trip?.status} /><span className="muted small mono right">{trip.reference}</span></div>
          <dl className="kv">
            <dt>Passenger</dt><dd>{trip.customer?.name}</dd>
            <dt>Pickup</dt><dd>{trip.pickup.address} · <a href={directions(trip.pickup.lat, trip.pickup.lng)} target="_blank" rel="noreferrer">Navigate</a></dd>
            <dt>Drop-off</dt><dd>{trip.dropoff.address} · <a href={directions(trip.dropoff.lat, trip.dropoff.lng)} target="_blank" rel="noreferrer">Navigate</a></dd>
            {trip.note && <><dt>Note</dt><dd>{trip.note}</dd></>}
            <dt>Estimate</dt><dd>{money(trip.estimate?.totalMinor, trip.estimate?.currency)} · cash</dd>
          </dl>
          {(NEXT[trip.trip?.status] ?? []).map((n) => (
            <button key={n.action} className="btn primary block" disabled={act.busy}
              onClick={() => act.run(async () => {
                const r = await api<{ booking: any }>("POST", `${base}/driver/trips/${trip.id}/${n.action}`);
                if (n.action === "complete") setFinished(r.booking);
                await refresh();
              })}>{n.label}</button>
          ))}
          <button className="btn danger" disabled={act.busy} onClick={() => {
            const reason = prompt("Why are you ending this trip? (e.g. passenger didn't show)");
            if (reason) go(() => api("POST", `${base}/driver/trips/${trip.id}/abort`, { reason }));
          }}>End trip early</button>
        </div>
      )}

      {finished && <CashCard base={base} booking={finished} onDone={() => setFinished(null)} />}

      {!trip && !offer && online && d.complianceIssues.length === 0 && (
        <div className="card empty">Waiting for ride offers. Keep this page open.</div>
      )}

      <Documents base={base} me={me} onChange={refresh} />
    </div>
  );
}

function OfferCard({ offer, busy, onAccept, onReject }: { offer: Offer; busy: boolean; onAccept: () => void; onReject: () => void }) {
  const [left, setLeft] = useState(0);
  useEffect(() => {
    const tick = () => setLeft(Math.max(0, Math.round((new Date(offer.expiresAt).getTime() - Date.now()) / 1000)));
    tick();
    const id = setInterval(tick, 500);
    return () => clearInterval(id);
  }, [offer.expiresAt]);
  return (
    <div className="card stack" style={{ borderColor: "var(--brand)", borderWidth: 2 }}>
      <div className="row"><h2 style={{ margin: 0 }}>New ride offer</h2><span className="badge warn right">{left}s</span></div>
      <dl className="kv">
        <dt>Pickup</dt><dd>{offer.pickupAddress} · {km(offer.distanceToPickupM)} away</dd>
        <dt>Drop-off</dt><dd>{offer.dropoffAddress}</dd>
        <dt>Passenger</dt><dd>{offer.customerFirstName}</dd>
        <dt>Estimate</dt><dd>{money(offer.estimatedTotalMinor, offer.currency)} · cash</dd>
      </dl>
      <div className="row">
        <button className="btn primary" style={{ flex: 1 }} disabled={busy || left === 0} onClick={onAccept}>Accept</button>
        <button className="btn" disabled={busy || left === 0} onClick={onReject}>Decline</button>
      </div>
    </div>
  );
}

function CashCard({ base, booking, onDone }: { base: string; booking: any; onDone: () => void }) {
  const total: number = booking.trip?.finalTotalMinor ?? 0;
  const [amount, setAmount] = useState((total / 100).toFixed(2));
  const [note, setNote] = useState("");
  const act = useAction();
  const minor = Math.round(Number(amount) * 100);
  return (
    <div className="card stack">
      <h2>Collect {money(total, booking.trip?.currency)} in cash</h2>
      <FareLines b={booking.trip?.finalBreakdown} title="Final fare" />
      <label className="field">Amount received (LKR)<input inputMode="decimal" value={amount} onChange={(e) => setAmount(e.target.value)} /></label>
      {minor !== total && <label className="field">Why is it different?<input required value={note} onChange={(e) => setNote(e.target.value)} /></label>}
      <ErrorBox error={act.error} />
      <button className="btn primary" disabled={act.busy} onClick={() => act.run(async () => {
        await api("POST", `${base}/driver/trips/${booking.id}/cash`, { amountMinor: minor, note });
        onDone();
      })}>Confirm cash received</button>
    </div>
  );
}

function Documents({ base, me, onChange }: { base: string; me: DriverMe; onChange: () => void }) {
  const [doc, setDoc] = useState({ docType: "driving_licence", number: "", expiresOn: "" });
  const act = useAction();
  return (
    <details className="card" open={me.driver.complianceIssues.length > 0}>
      <summary><strong>My documents</strong></summary>
      <div className="stack" style={{ marginTop: 12 }}>
        {me.documents.map((x) => (
          <div key={x.id} className="row small">
            <span>{x.docType.replaceAll("_", " ")}</span><span className="mono muted">{x.number}</span>
            {x.expiresOn && <span className="muted">exp. {x.expiresOn}</span>}<Badge value={x.status} />
            {x.reviewNote && <span className="muted">“{x.reviewNote}”</span>}
          </div>
        ))}
        <form className="fields" onSubmit={(e) => { e.preventDefault(); act.run(async () => {
          await api("POST", `${base}/driver/documents`, { ...doc, expiresOn: doc.expiresOn || null });
          setDoc({ ...doc, number: "", expiresOn: "" }); onChange();
        }); }}>
          <label className="field">Document<select value={doc.docType} onChange={(e) => setDoc({ ...doc, docType: e.target.value })}>{DOC_TYPES.map((x) => <option key={x} value={x}>{x.replaceAll("_", " ")}</option>)}</select></label>
          <label className="field">Number<input required value={doc.number} onChange={(e) => setDoc({ ...doc, number: e.target.value })} /></label>
          <label className="field">Expires<input type="date" value={doc.expiresOn} onChange={(e) => setDoc({ ...doc, expiresOn: e.target.value })} /></label>
          <div style={{ alignSelf: "end" }}><button className="btn" disabled={act.busy}>Submit</button></div>
        </form>
        <ErrorBox error={act.error} />
      </div>
    </details>
  );
}

export default function Page() {
  return (
    <>
      <TopBar title="Driver"><SignOut /></TopBar>
      <main className="page narrow" style={{ maxWidth: 560 }}>
        <Suspense><DriverApp /></Suspense>
      </main>
    </>
  );
}
