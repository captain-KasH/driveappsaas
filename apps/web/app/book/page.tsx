"use client";

import dynamic from "next/dynamic";
import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { ago, api, ApiError, km, Me, mins, money, newIdempotencyKey, token, when } from "@/lib/api";
import { Badge, ErrorBox, Spinner, TopBar, useAction, usePoll } from "@/components/ui";
import PlaceInput from "@/components/PlaceInput";
import { Marker, Place, reverseGeocode, MAP_COLORS } from "@/components/MapView";
import FareLines, { Breakdown } from "@/components/FareLines";

const MapView = dynamic(() => import("@/components/MapView"), { ssr: false, loading: () => <div className="map" /> });

type Brand = { tenant: { id: string; name: string }; brand: PublicBrand; categories: { id: string; name: string; seats: number }[] };
type PublicBrand = { name: string; color: string; onColor: string; logoUrl: string | null; tagline: string | null; supportPhone: string | null; supportEmail: string | null; websiteUrl: string | null; termsUrl: string | null; privacyUrl: string | null };

const STATUS_TEXT: Record<string, string> = {
  scheduled: "Scheduled — we'll start finding a driver 15 minutes before pickup.",
  searching: "Finding you a driver…",
  driver_assigned: "Driver assigned and getting ready.",
  en_route: "Your driver is on the way.",
  arrived: "Your driver has arrived.",
  in_progress: "On the way to your destination.",
  completed: "Trip complete. Thank you for riding!",
  cancelled: "This booking was cancelled.",
  expired: "Sorry — no driver accepted in time. Please try again.",
};

function BookApp() {
  const slug = useSearchParams().get("b") ?? "";
  const [brand, setBrand] = useState<Brand | null>(null);
  const [brandErr, setBrandErr] = useState<unknown>(null);
  const [member, setMember] = useState<"unknown" | "yes" | "no">("unknown");

  useEffect(() => {
    if (!slug) return;
    api<Brand>("GET", `/v1/public/tenants/${encodeURIComponent(slug)}`).then(setBrand).catch(setBrandErr);
  }, [slug]);

  const checkMember = useCallback(async (b: Brand) => {
    if (!token.get()) return setMember("no");
    try {
      const me = await api<Me>("GET", "/v1/me");
      setMember(me.memberships.some((m) => m.tenantId === b.tenant.id && m.role === "passenger") ? "yes" : "no");
    } catch { setMember("no"); }
  }, []);
  useEffect(() => { if (brand) checkMember(brand); }, [brand, checkMember]);

  if (!slug) return <div className="alert error">This booking link is incomplete. Ask your operator for their Taxios booking link.</div>;
  if (brandErr) return <div className="alert error">{brandErr instanceof ApiError && brandErr.status === 404 ? "This operator isn't taking bookings on Taxios right now." : String(brandErr)}</div>;
  if (!brand || member === "unknown") return <Spinner />;

  const b = brand.brand;
  return (
    <div className="stack" style={{ ["--brand" as string]: b.color, ["--brand-ink" as string]: b.onColor, ["--accent" as string]: b.color }}>
      <div className="row">
        {b.logoUrl && <img src={b.logoUrl} alt="" style={{ height: 44, borderRadius: 8 }} />}
        <div><h1 style={{ margin: 0 }}>Book with {brand.tenant.name}</h1>{b.tagline && <p className="muted" style={{ margin: 0 }}>{b.tagline}</p>}</div>
      </div>
      {member === "no" ? <Join slug={slug} brand={brand} onJoined={() => setMember("yes")} /> : <Booking brand={brand} />}
      {(b.supportPhone || b.supportEmail || b.termsUrl || b.privacyUrl || b.websiteUrl) && (
        <p className="muted small">
          {b.supportPhone && <>Call <a href={`tel:${b.supportPhone}`}>{b.supportPhone}</a> · </>}
          {b.supportEmail && <><a href={`mailto:${b.supportEmail}`}>{b.supportEmail}</a> · </>}
          {b.websiteUrl && <><a href={b.websiteUrl} rel="noopener">Website</a> · </>}
          {b.termsUrl && <><a href={b.termsUrl} rel="noopener">Terms</a> · </>}
          {b.privacyUrl && <a href={b.privacyUrl} rel="noopener">Privacy</a>}
        </p>
      )}
    </div>
  );
}

function Join({ slug, brand, onJoined }: { slug: string; brand: Brand; onJoined: () => void }) {
  const [mode, setMode] = useState<"register" | "login">("register");
  const [f, setF] = useState({ fullName: "", email: "", phone: "", password: "" });
  const act = useAction();
  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    act.run(async () => {
      if (mode === "login") {
        const r = await api<{ token: string }>("POST", "/v1/auth/login", { email: f.email, password: f.password });
        token.set(r.token);
        const me = await api<Me>("GET", "/v1/me");
        if (me.memberships.some((m) => m.tenantId === brand.tenant.id && m.role === "passenger")) return onJoined();
        setMode("register");
        throw new Error(`You're signed in. Add your name and phone once to ride with ${brand.tenant.name}.`);
      }
      const r = await api<{ token: string }>("POST", `/v1/public/tenants/${encodeURIComponent(slug)}/passengers`, f);
      token.set(r.token);
      onJoined();
    });
  };
  return (
    <form className="card stack narrow" onSubmit={submit}>
      <div className="tabs" role="tablist">
        <button type="button" role="tab" className="tab" aria-selected={mode === "register"} onClick={() => setMode("register")}>Create account</button>
        <button type="button" role="tab" className="tab" aria-selected={mode === "login"} onClick={() => setMode("login")}>Sign in</button>
      </div>
      {mode === "register" && <>
        <label className="field">Full name<input required autoComplete="name" value={f.fullName} onChange={(e) => setF({ ...f, fullName: e.target.value })} /></label>
        <label className="field">Mobile number<input required inputMode="tel" autoComplete="tel" placeholder="077 123 4567" value={f.phone} onChange={(e) => setF({ ...f, phone: e.target.value })} /></label>
      </>}
      <label className="field">Email<input required type="email" autoComplete="email" value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></label>
      <label className="field">Password{mode === "register" && " (10+ characters)"}<input required type="password" minLength={mode === "register" ? 10 : undefined}
        autoComplete={mode === "register" ? "new-password" : "current-password"} value={f.password} onChange={(e) => setF({ ...f, password: e.target.value })} /></label>
      <ErrorBox error={act.error} />
      <button className="btn primary block" disabled={act.busy}>{mode === "register" ? "Create account" : "Sign in"}</button>
      {mode === "register" && <p className="muted small">Already have a Taxios account? Use the same email and password — we'll link it.</p>}
    </form>
  );
}

function Booking({ brand }: { brand: Brand }) {
  const base = `/v1/tenants/${brand.tenant.id}`;
  const [pickup, setPickup] = useState<Place | null>(null);
  const [dropoff, setDropoff] = useState<Place | null>(null);
  const [focus, setFocus] = useState<"pickup" | "dropoff">("pickup");
  const [categoryId, setCategoryId] = useState(brand.categories[0]?.id ?? "");
  const [later, setLater] = useState("");
  const [note, setNote] = useState("");
  const [quote, setQuote] = useState<any>(null);
  const [key, setKey] = useState(newIdempotencyKey);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [history, setHistory] = useState<any[] | null>(null);
  const act = useAction();

  const loadHistory = useCallback(() => api<{ bookings: any[] }>("GET", `${base}/my/bookings`).then((r) => {
    setHistory(r.bookings);
    const live = r.bookings.find((b) => ["scheduled", "searching", "assigned"].includes(b.status));
    if (live) setActiveId((cur) => cur ?? live.id);
  }).catch(() => {}), [base]);
  useEffect(() => { loadHistory(); }, [loadHistory]);

  const pick = async (p: { lat: number; lng: number }) => {
    const place = { ...p, address: await reverseGeocode(p) };
    setQuote(null);
    if (focus === "pickup") { setPickup(place); setFocus("dropoff"); } else setDropoff(place);
  };
  const locate = () => navigator.geolocation?.getCurrentPosition(
    (p) => pick({ lat: p.coords.latitude, lng: p.coords.longitude }).then(() => setFocus("dropoff")),
    () => act.setError(new Error("Couldn't get your location. Search or tap the map instead.")),
  );

  const getQuote = () => act.run(async () => {
    if (!pickup || !dropoff) throw new Error("Set your pickup and drop-off first.");
    const r = await api<{ quote: any }>("POST", `${base}/quotes`, {
      pickup, dropoff, categoryId, ...(later && { scheduledFor: new Date(later).toISOString() }),
    });
    setQuote(r.quote);
    setKey(newIdempotencyKey());
  });
  const book = () => act.run(async () => {
    const r = await api<{ booking: any }>("POST", `${base}/my/bookings`, { quoteId: quote.id, paymentMethod: "cash", note }, { "Idempotency-Key": key });
    setActiveId(r.booking.id);
    setQuote(null); setPickup(null); setDropoff(null); setNote(""); setLater("");
    loadHistory();
  });

  if (activeId) return <ActiveBooking base={base} id={activeId} onClose={() => { setActiveId(null); loadHistory(); }} />;

  const markers: Marker[] = [
    ...(pickup ? [{ ...pickup, label: "Pickup", color: MAP_COLORS.pickup }] : []),
    ...(dropoff ? [{ ...dropoff, label: "Drop-off", color: MAP_COLORS.dropoff }] : []),
  ];
  return (
    <div className="split">
      <div className="card stack">
        <PlaceInput label="Pickup" value={pickup} onChange={(p) => { setPickup(p); setQuote(null); setFocus("dropoff"); }} active={focus === "pickup"} onFocus={() => setFocus("pickup")} />
        <button type="button" className="linkbtn small" style={{ alignSelf: "flex-start" }} onClick={() => { setFocus("pickup"); locate(); }}>Use my current location</button>
        <PlaceInput label="Drop-off" value={dropoff} onChange={(p) => { setDropoff(p); setQuote(null); }} active={focus === "dropoff"} onFocus={() => setFocus("dropoff")} />
        <div className="fields">
          <label className="field">Vehicle<select value={categoryId} onChange={(e) => { setCategoryId(e.target.value); setQuote(null); }}>
            {brand.categories.map((c) => <option key={c.id} value={c.id}>{c.name} · {c.seats} seats</option>)}</select></label>
          <label className="field">When<input type="datetime-local" value={later} onChange={(e) => { setLater(e.target.value); setQuote(null); }} aria-describedby="when-help" /></label>
        </div>
        <span id="when-help" className="muted small">Leave “When” empty to ride now.</span>
        {!quote ? (
          <button className="btn primary" disabled={act.busy || !pickup || !dropoff} onClick={getQuote}>See fare</button>
        ) : (
          <div className="stack">
            <FareLines b={quote.estimate as Breakdown} title={`Estimated fare · about ${km(quote.estimatedDistanceM)}, ${mins(quote.estimatedDurationS)}`} />
            <p className="muted small">This is an estimate. Your final fare is calculated from the actual trip using the same rates. Pay the driver in cash.</p>
            <label className="field">Note for the driver (optional)<input maxLength={500} value={note} onChange={(e) => setNote(e.target.value)} /></label>
            <button className="btn primary" disabled={act.busy} onClick={book}>Book · {money(quote.estimate.totalMinor, quote.estimate.currency)} est.</button>
          </div>
        )}
        <ErrorBox error={act.error} />
      </div>
      <div className="stack">
        <MapView markers={markers} onPick={pick} fitKey={markers.map((m) => m.lat).join()} />
        <p className="muted small">Tap the map to set the {focus === "pickup" ? "pickup" : "drop-off"}.</p>
        {history && history.length > 0 && (
          <div className="card">
            <h3>Your rides</h3>
            {history.slice(0, 8).map((b) => (
              <button key={b.id} className="btn small block" style={{ justifyContent: "space-between", border: 0 }} onClick={() => setActiveId(b.id)}>
                <span>{when(b.createdAt)} · {b.dropoffAddress.split(",")[0]}</span><Badge value={b.status} />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function ActiveBooking({ base, id, onClose }: { base: string; id: string; onClose: () => void }) {
  const [b, setB] = useState<any>(null);
  const [receipt, setReceipt] = useState<any>(null);
  const act = useAction();
  const load = useCallback(() => api<{ booking: any }>("GET", `${base}/my/bookings/${id}`).then((r) => setB(r.booking)).catch(act.setError), [base, id, act.setError]);
  usePoll(load, 4000);
  useEffect(() => {
    if (b?.status === "completed" && !receipt) api<{ receipt: any }>("GET", `${base}/my/bookings/${id}/receipt`).then((r) => setReceipt(r.receipt)).catch(() => {});
  }, [b?.status, base, id, receipt]);

  const markers = useMemo<Marker[]>(() => {
    if (!b) return [];
    const m: Marker[] = [{ ...b.pickup, label: "Pickup", color: MAP_COLORS.pickup }, { ...b.dropoff, label: "Drop-off", color: MAP_COLORS.dropoff }];
    if (b.driver?.lat != null && b.driver.locationFresh) m.push({ lat: b.driver.lat, lng: b.driver.lng, label: `${b.driver.firstName} (driver)`, color: MAP_COLORS.driver });
    return m;
  }, [b]);

  if (!b) return <Spinner />;
  const state = b.status === "assigned" ? b.trip?.status : b.status;
  const cancellable = ["scheduled", "searching"].includes(b.status) || (b.status === "assigned" && b.trip?.status !== "in_progress");
  return (
    <div className="split">
      <div className="card stack">
        <div className="row"><h2 style={{ margin: 0 }} className="mono">{b.reference}</h2><Badge value={state} /></div>
        <p>{STATUS_TEXT[state] ?? state}</p>
        {b.driver && b.status === "assigned" && (
          <div className="card" style={{ background: "var(--surface-2)" }}>
            <strong>{b.driver.firstName}</strong> · {b.driver.vehicleDescription} · <span className="mono">{b.driver.vehiclePlate}</span>
            <div className="muted small">{b.driver.locationFresh ? `Location updated ${ago(b.driver.locationAt)}` : "Driver location not available right now"}</div>
          </div>
        )}
        <dl className="kv">
          <dt>Pickup</dt><dd>{b.pickup.address}</dd>
          <dt>Drop-off</dt><dd>{b.dropoff.address}</dd>
          {b.scheduledFor && <><dt>Pickup time</dt><dd>{when(b.scheduledFor)}</dd></>}
        </dl>
        {receipt ? <Receipt r={receipt} base={base} id={id} /> : <FareLines b={b.estimate} title="Estimated fare" />}
        <ErrorBox error={act.error} />
        <div className="row">
          {cancellable && <button className="btn danger" disabled={act.busy} onClick={() => {
            if (confirm("Cancel this booking?")) act.run(async () => { await api("POST", `${base}/my/bookings/${id}/cancel`, { reason: "cancelled by passenger" }); await load(); });
          }}>Cancel booking</button>}
          <button className="btn" onClick={onClose}>{cancellable ? "Back" : "Book another ride"}</button>
        </div>
      </div>
      <MapView className="map tall" markers={markers} fitKey={id} />
    </div>
  );
}

function Receipt({ r, base, id }: { r: any; base: string; id: string }) {
  const [stars, setStars] = useState(0);
  const [comment, setComment] = useState("");
  const [rated, setRated] = useState(false);
  const act = useAction();
  return (
    <div className="stack">
      <h3>Receipt · {r.business}</h3>
      <div className="muted small">{km(r.distanceM)} · {mins(r.durationS)} · {r.distanceSource === "gps_track" ? "distance measured by GPS" : "estimated distance"}</div>
      <FareLines b={r.fare} />
      <div className="small">Payment: {r.paymentStatus === "paid_cash" ? `paid in cash (${money(r.paidMinor)})` : "cash, not yet confirmed by the driver"}</div>
      {!rated ? (
        <div className="stack">
          <div className="row" role="radiogroup" aria-label="Rate your trip">
            {[1, 2, 3, 4, 5].map((n) => (
              <button key={n} type="button" role="radio" aria-checked={stars === n} className="btn small" onClick={() => setStars(n)}
                style={stars >= n ? { background: "var(--brand)", color: "var(--brand-ink)" } : undefined}>★ {n}</button>
            ))}
          </div>
          {stars > 0 && <>
            <input placeholder="Anything to add? (optional)" maxLength={1000} value={comment} onChange={(e) => setComment(e.target.value)} />
            <button className="btn" disabled={act.busy} onClick={() => act.run(async () => {
              try { await api("POST", `${base}/my/bookings/${id}/rating`, { stars, comment }); }
              catch (e) { if (!(e instanceof ApiError && e.code === "already_rated")) throw e; }
              setRated(true);
            })}>Send rating</button>
          </>}
          <ErrorBox error={act.error} />
        </div>
      ) : <div className="alert ok">Thanks for your feedback.</div>}
    </div>
  );
}

export default function Page() {
  return (
    <>
      <TopBar title="Book a ride" />
      <main className="page">
        <Suspense><BookApp /></Suspense>
      </main>
    </>
  );
}
