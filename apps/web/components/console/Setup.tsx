"use client";

import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useState } from "react";
import { api, money } from "@/lib/api";
import { ErrorBox, Spinner, useAction } from "@/components/ui";
import { circlePolygon, Pt } from "@/components/MapView";
import { Category, Ctx, Zone, can, fromMinor, toMinor } from "./ctx";

const MapView = dynamic(() => import("@/components/MapView"), { ssr: false, loading: () => <div className="map" /> });

type Fare = {
  id: string; zoneId: string; categoryId: string; version: number; currency: string; baseMinor: number; perKmMinor: number;
  perMinuteMinor: number; minimumMinor: number; bookingFeeMinor: number; scheduledFeeMinor: number; waitingPerMinMinor: number;
  freeWaitingMinutes: number; nightSurchargeBps: number; nightStart: string | null; nightEnd: string | null; taxBps: number; roundToMinor: number;
};

const FARE_FIELDS: [keyof Fare, string, "money" | "int" | "pct"][] = [
  ["baseMinor", "Base fare", "money"], ["perKmMinor", "Per km", "money"], ["perMinuteMinor", "Per minute", "money"],
  ["minimumMinor", "Minimum fare", "money"], ["bookingFeeMinor", "Booking fee", "money"], ["scheduledFeeMinor", "Scheduled ride fee", "money"],
  ["waitingPerMinMinor", "Waiting per minute", "money"], ["freeWaitingMinutes", "Free waiting (min)", "int"],
  ["nightSurchargeBps", "Night surcharge %", "pct"], ["taxBps", "Tax %", "pct"],
];

export default function Setup({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const editable = can(ctx.role, "pricing.manage");
  const [zones, setZones] = useState<Zone[] | null>(null);
  const [cats, setCats] = useState<Category[] | null>(null);
  const [fares, setFares] = useState<Fare[] | null>(null);
  const [loadErr, setLoadErr] = useState<unknown>(null);

  const load = useCallback(() => Promise.all([
    api<{ zones: Zone[] }>("GET", `${base}/zones`).then((r) => setZones(r.zones)),
    api<{ categories: Category[] }>("GET", `${base}/vehicle-categories`).then((r) => setCats(r.categories)),
    api<{ fareRules: Fare[] }>("GET", `${base}/fare-rules`).then((r) => setFares(r.fareRules)),
  ]).catch(setLoadErr), [base]);
  useEffect(() => { load(); }, [load]);

  const finish = useAction();
  if (!zones || !cats || !fares) return loadErr ? <ErrorBox error={loadErr} /> : <Spinner />;

  const steps = [
    ["Draw a service zone", zones.length > 0],
    ["Add a vehicle category", cats.length > 0],
    ["Publish a fare rule", fares.length > 0],
  ] as const;

  return (
    <div className="stack">
      {ctx.tenant.status === "onboarding" && (
        <div className="card stack">
          <h2>Get ready to take bookings</h2>
          <ol className="stack" style={{ margin: 0, paddingLeft: 20 }}>
            {steps.map(([label, done]) => <li key={label}>{done ? "✅" : "⬜"} {label}</li>)}
          </ol>
          <ErrorBox error={finish.error} />
          {ctx.role === "owner" && (
            <div>
              <button className="btn primary" disabled={finish.busy || !steps.every((s) => s[1])}
                onClick={() => finish.run(async () => { await api("POST", `${base}/onboarding/complete`); await ctx.reload(); })}>
                Go live
              </button>
              <span className="muted small"> Passengers can book once you're live.</span>
            </div>
          )}
        </div>
      )}
      <ZoneCard base={base} zones={zones} editable={editable} onChange={load} />
      <CategoryCard base={base} cats={cats} editable={editable} onChange={load} />
      <FareCard base={base} zones={zones} cats={cats} fares={fares} editable={editable} onChange={load} />
    </div>
  );
}

function ZoneCard({ base, zones, editable, onChange }: { base: string; zones: Zone[]; editable: boolean; onChange: () => void }) {
  const [name, setName] = useState("");
  const [centre, setCentre] = useState<Pt | null>(null);
  const [radius, setRadius] = useState(8);
  const [geojson, setGeojson] = useState("");
  const act = useAction();
  const draft = useMemo(() => (centre ? circlePolygon(centre, radius) : null), [centre, radius]);
  const polygons = useMemo(() => [...zones.map((z) => z.area), ...(draft ? [draft] : [])], [zones, draft]);

  const save = (e: React.FormEvent) => {
    e.preventDefault();
    act.run(async () => {
      const area = geojson.trim() ? JSON.parse(geojson) : draft;
      if (!area) throw new Error("Tap the map to set the zone centre, or paste a GeoJSON polygon.");
      await api("POST", `${base}/zones`, { name, area });
      setName(""); setCentre(null); setGeojson("");
      onChange();
    });
  };

  return (
    <div className="card stack">
      <h2>Service zones</h2>
      <p className="muted small">Pickups must fall inside an active zone. {editable && "Tap the map to place a zone centre, then set its radius."}</p>
      <MapView className="map" polygons={polygons} markers={centre ? [{ ...centre, label: "Zone centre", color: "#8a5a00" }] : []}
        onPick={editable ? setCentre : undefined} fitKey={String(zones.length)} />
      {zones.length > 0 && <div className="row small">{zones.map((z) => <span key={z.id} className="badge">{z.name}</span>)}</div>}
      {editable && (
        <form className="stack" onSubmit={save}>
          <div className="fields">
            <label className="field">Zone name<input required value={name} onChange={(e) => setName(e.target.value)} placeholder="Colombo" /></label>
            <label className="field">Radius: {radius} km<input type="range" min={1} max={40} value={radius} onChange={(e) => setRadius(Number(e.target.value))} /></label>
          </div>
          <details className="small">
            <summary>Or paste an exact boundary (GeoJSON Polygon)</summary>
            <textarea className="mono" value={geojson} onChange={(e) => setGeojson(e.target.value)} placeholder='{"type":"Polygon","coordinates":[[[79.8,6.8],...]]}' />
          </details>
          <ErrorBox error={act.error} />
          <div><button className="btn primary" disabled={act.busy}>Save zone</button></div>
        </form>
      )}
    </div>
  );
}

function CategoryCard({ base, cats, editable, onChange }: { base: string; cats: Category[]; editable: boolean; onChange: () => void }) {
  const [f, setF] = useState({ code: "", name: "", seats: 4 });
  const act = useAction();
  return (
    <div className="card stack">
      <h2>Vehicle categories</h2>
      {cats.length === 0 ? <div className="muted small">None yet — e.g. Tuk (3 seats), Car (4), Van (8).</div> : (
        <div className="row">{cats.map((c) => <span key={c.id} className="badge">{c.name} · {c.seats} seats</span>)}</div>
      )}
      {editable && (
        <form className="fields" onSubmit={(e) => { e.preventDefault(); act.run(async () => { await api("POST", `${base}/vehicle-categories`, f); setF({ code: "", name: "", seats: 4 }); onChange(); }); }}>
          <label className="field">Name<input required value={f.name} onChange={(e) => setF({ ...f, name: e.target.value, code: f.code || "" })} placeholder="Car" /></label>
          <label className="field">Code<input required pattern="[a-z0-9_]{2,32}" title="2–32 lowercase letters, digits, _" value={f.code} placeholder="car"
            onChange={(e) => setF({ ...f, code: e.target.value.toLowerCase() })} /></label>
          <label className="field">Seats<input type="number" min={1} max={60} required value={f.seats} onChange={(e) => setF({ ...f, seats: Number(e.target.value) })} /></label>
          <div style={{ alignSelf: "end" }}><button className="btn primary" disabled={act.busy}>Add category</button></div>
          <div style={{ gridColumn: "1/-1" }}><ErrorBox error={act.error} /></div>
        </form>
      )}
    </div>
  );
}

function FareCard({ base, zones, cats, fares, editable, onChange }: {
  base: string; zones: Zone[]; cats: Category[]; fares: Fare[]; editable: boolean; onChange: () => void;
}) {
  const blank = { zoneId: "", categoryId: "", nightStart: "", nightEnd: "", roundTo: "1.00" } as Record<string, string>;
  const [f, setF] = useState<Record<string, string>>(blank);
  const act = useAction();
  const zoneName = (id: string) => zones.find((z) => z.id === id)?.name ?? "?";
  const catName = (id: string) => cats.find((c) => c.id === id)?.name ?? "?";

  const edit = (r: Fare) => {
    const v: Record<string, string> = { zoneId: r.zoneId, categoryId: r.categoryId, nightStart: r.nightStart ?? "", nightEnd: r.nightEnd ?? "", roundTo: fromMinor(r.roundToMinor) };
    for (const [k, , kind] of FARE_FIELDS) {
      const n = r[k] as number;
      v[k] = kind === "money" ? fromMinor(n) : kind === "pct" ? String(n / 100) : String(n);
    }
    setF(v);
  };

  const publish = (e: React.FormEvent) => {
    e.preventDefault();
    act.run(async () => {
      const body: Record<string, unknown> = { zoneId: f.zoneId, categoryId: f.categoryId, roundToMinor: toMinor(f.roundTo || "1") };
      for (const [k, , kind] of FARE_FIELDS) {
        const v = f[k] ?? "";
        body[k] = kind === "money" ? toMinor(v || "0") : kind === "pct" ? Math.round(Number(v || 0) * 100) : Number(v || 0);
      }
      if (f.nightStart && f.nightEnd) { body.nightStart = f.nightStart; body.nightEnd = f.nightEnd; }
      await api("POST", `${base}/fare-rules`, body);
      setF(blank);
      onChange();
    });
  };

  return (
    <div className="card stack">
      <h2>Fare rules</h2>
      <p className="muted small">Publishing creates a new version. Existing bookings keep the version they were quoted with.</p>
      {fares.length > 0 && (
        <table className="list">
          <thead><tr><th>Zone · category</th><th>Base</th><th>Per km</th><th>Per min</th><th>Minimum</th><th>Version</th><th></th></tr></thead>
          <tbody>
            {fares.map((r) => (
              <tr key={r.id}>
                <td data-label="Rule">{zoneName(r.zoneId)} · {catName(r.categoryId)}</td>
                <td data-label="Base">{money(r.baseMinor, r.currency)}</td>
                <td data-label="Per km">{money(r.perKmMinor, r.currency)}</td>
                <td data-label="Per min">{money(r.perMinuteMinor, r.currency)}</td>
                <td data-label="Minimum">{money(r.minimumMinor, r.currency)}</td>
                <td data-label="Version">v{r.version}</td>
                <td>{editable && <button className="btn small" onClick={() => edit(r)}>Edit</button>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {editable && (zones.length === 0 || cats.length === 0 ? <div className="muted small">Add a zone and a category first.</div> : (
        <form className="stack" onSubmit={publish}>
          <div className="fields">
            <label className="field">Zone<select required value={f.zoneId} onChange={(e) => setF({ ...f, zoneId: e.target.value })}>
              <option value="">Choose…</option>{zones.map((z) => <option key={z.id} value={z.id}>{z.name}</option>)}</select></label>
            <label className="field">Category<select required value={f.categoryId} onChange={(e) => setF({ ...f, categoryId: e.target.value })}>
              <option value="">Choose…</option>{cats.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label>
            {FARE_FIELDS.map(([k, label, kind]) => (
              <label className="field" key={k}>{label}{kind === "money" && " (LKR)"}
                <input inputMode="decimal" pattern={kind === "int" ? "\\d+" : "\\d+(\\.\\d{1,2})?"} value={f[k] ?? ""} placeholder="0"
                  onChange={(e) => setF({ ...f, [k]: e.target.value })} />
              </label>
            ))}
            <label className="field">Night from<input type="time" value={f.nightStart} onChange={(e) => setF({ ...f, nightStart: e.target.value })} /></label>
            <label className="field">Night until<input type="time" value={f.nightEnd} onChange={(e) => setF({ ...f, nightEnd: e.target.value })} /></label>
            <label className="field">Round total to (LKR)<input inputMode="decimal" value={f.roundTo} onChange={(e) => setF({ ...f, roundTo: e.target.value })} /></label>
          </div>
          <ErrorBox error={act.error} />
          <div><button className="btn primary" disabled={act.busy}>Publish fare rule</button></div>
        </form>
      ))}
    </div>
  );
}
