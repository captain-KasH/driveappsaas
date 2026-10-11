"use client";

import { useCallback, useEffect, useState } from "react";
import { ago, api, openFile } from "@/lib/api";
import FilePick from "@/components/FilePick";
import { Badge, ErrorBox, Spinner, useAction } from "@/components/ui";
import { CopyLink } from "@/components/CopyLink";
import { Ctx, Driver, ISSUE_TEXT, Vehicle, can } from "./ctx";

type Doc = { id: string; docType: string; number: string; expiresOn: string | null; status: string; reviewNote: string | null; fileId: string | null };
const DOC_TYPES = ["driving_licence", "national_id", "vehicle_insurance", "revenue_licence", "police_report"];

export default function Drivers({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const manage = can(ctx.role, "fleet.manage");
  const [drivers, setDrivers] = useState<Driver[] | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback(() => api<{ drivers: Driver[] }>("GET", `${base}/drivers`).then((r) => setDrivers(r.drivers)).catch(setError), [base]);
  useEffect(() => { load(); }, [load]);

  return (
    <div className="split">
      <div className="stack">
        {manage && <AddDriver base={base} onAdded={load} />}
        <div className="card">
          <h2>Drivers</h2>
          <ErrorBox error={error} />
          {!drivers ? <Spinner /> : drivers.length === 0 ? <div className="empty">No drivers yet.</div> : (
            <table className="list">
              <tbody>
                {drivers.map((d) => (
                  <tr key={d.id} className="clickable" onClick={() => setSelected(d.id)} aria-selected={selected === d.id}>
                    <td data-label="Driver">
                      <strong>{d.fullName}</strong>
                      <div className="muted small">{d.phone}{!d.hasAccount && " · invite not accepted"}</div>
                      {d.complianceIssues.length > 0 && <div className="small" style={{ color: "var(--warn)" }}>{d.complianceIssues.map((i) => ISSUE_TEXT[i] ?? i).join(", ")}</div>}
                    </td>
                    <td data-label="Status"><Badge value={d.status} /> <Badge value={d.availability} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
      <div>{selected ? <DriverDetail key={selected} base={base} id={selected} manage={manage} onChange={load} /> : <div className="card empty">Select a driver to review documents and approve.</div>}</div>
    </div>
  );
}

function AddDriver({ base, onAdded }: { base: string; onAdded: () => void }) {
  const [f, setF] = useState({ fullName: "", phone: "", email: "" });
  const [link, setLink] = useState("");
  const act = useAction();
  return (
    <form className="card stack" onSubmit={(e) => {
      e.preventDefault();
      act.run(async () => {
        const r = await api<{ invitation: { token: string } }>("POST", `${base}/drivers`, f);
        setLink(`${location.origin}/invite/?token=${r.invitation.token}`);
        setF({ fullName: "", phone: "", email: "" });
        onAdded();
      });
    }}>
      <h2>Add a driver</h2>
      <div className="fields">
        <label className="field">Full name<input required value={f.fullName} onChange={(e) => setF({ ...f, fullName: e.target.value })} /></label>
        <label className="field">Phone<input required inputMode="tel" placeholder="077 123 4567" value={f.phone} onChange={(e) => setF({ ...f, phone: e.target.value })} /></label>
        <label className="field">Email<input type="email" required value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></label>
      </div>
      <ErrorBox error={act.error} />
      <div><button className="btn primary" disabled={act.busy}>Add and create invite</button></div>
      {link && <div className="alert ok stack"><span>Send this one-time link to the driver (expires in 7 days):</span><CopyLink value={link} /></div>}
    </form>
  );
}

function DriverDetail({ base, id, manage, onChange }: { base: string; id: string; manage: boolean; onChange: () => void }) {
  const [d, setD] = useState<{ driver: Driver; documents: Doc[]; vehicle: Vehicle | null } | null>(null);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [doc, setDoc] = useState({ docType: "driving_licence", number: "", expiresOn: "" });
  const [docFile, setDocFile] = useState<{ id: string; fileName: string } | null>(null);
  const act = useAction();
  const load = useCallback(() => api("GET", `${base}/drivers/${id}`).then(setD).catch(act.setError), [base, id, act.setError]);
  useEffect(() => { load(); if (manage) api<{ vehicles: Vehicle[] }>("GET", `${base}/vehicles`).then((r) => setVehicles(r.vehicles)).catch(() => {}); }, [load, base, manage]);

  const go = (fn: () => Promise<unknown>) => act.run(async () => { await fn(); await load(); onChange(); });
  if (!d) return <div className="card"><Spinner /></div>;
  const { driver, documents, vehicle } = d;

  return (
    <div className="card stack">
      <div className="row">
        <h2 style={{ margin: 0 }}>{driver.fullName}</h2>
        <Badge value={driver.status} /><Badge value={driver.availability} />
      </div>
      <dl className="kv">
        <dt>Phone</dt><dd>{driver.phone}</dd>
        <dt>Email</dt><dd>{driver.email}</dd>
        <dt>Location</dt><dd>{driver.locationAt ? <>{ago(driver.locationAt)} {driver.locationFresh ? <Badge value="live" /> : <span className="badge bad">stale</span>}</> : "never shared"}</dd>
        <dt>Vehicle</dt><dd>{vehicle ? `${vehicle.plate} · ${vehicle.colour} ${vehicle.make} ${vehicle.model}` : "none"} {vehicle && <Badge value={vehicle.status} />}</dd>
        <dt>Can drive?</dt><dd>{driver.complianceIssues.length === 0 ? <span className="badge ok">eligible</span> : driver.complianceIssues.map((i) => ISSUE_TEXT[i] ?? i).join(", ")}</dd>
      </dl>
      <ErrorBox error={act.error} />

      <h3>Documents</h3>
      <p className="muted small">Required to go online: driving licence, national ID and vehicle insurance — approved and not expired.</p>
      {documents.length === 0 ? <div className="muted small">No documents submitted yet.</div> : (
        <table className="list">
          <tbody>
            {documents.map((x) => (
              <tr key={x.id}>
                <td data-label="Document"><strong>{x.docType.replaceAll("_", " ")}</strong><div className="muted small mono">{x.number}</div></td>
                <td data-label="Expires">{x.expiresOn ?? "—"}{x.expiresOn && x.expiresOn < new Date().toISOString().slice(0, 10) && <span className="badge bad"> expired</span>}</td>
                <td data-label="Status"><Badge value={x.status} /></td>
                <td data-label="File">{x.fileId ? <button className="linkbtn" onClick={() => openFile(`${base}/files/${x.fileId}`).catch(act.setError)}>View</button> : <span className="muted small">no scan</span>}</td>
                <td>{manage && x.status !== "approved" && (
                  <div className="row" style={{ gap: 6 }}>
                    <button className="btn small" disabled={act.busy} onClick={() => go(() => api("POST", `${base}/drivers/${id}/documents/${x.id}/review`, { decision: "approved" }))}>Approve</button>
                    <button className="btn small danger" disabled={act.busy} onClick={() => {
                      const note = prompt("Why is this document rejected? (the driver will see this)");
                      if (note) go(() => api("POST", `${base}/drivers/${id}/documents/${x.id}/review`, { decision: "rejected", note }));
                    }}>Reject</button>
                  </div>
                )}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {manage && (
        <form className="fields" onSubmit={(e) => { e.preventDefault(); go(() => api("POST", `${base}/drivers/${id}/documents`, { ...doc, expiresOn: doc.expiresOn || null, fileId: docFile?.id ?? null })); setDoc({ ...doc, number: "", expiresOn: "" }); setDocFile(null); }}>
          <FilePick key={docFile?.id ?? "empty"} base={base} value={docFile} onChange={setDocFile} label="Scan (optional)" />
          <label className="field">Type<select value={doc.docType} onChange={(e) => setDoc({ ...doc, docType: e.target.value })}>{DOC_TYPES.map((t) => <option key={t} value={t}>{t.replaceAll("_", " ")}</option>)}</select></label>
          <label className="field">Number<input required value={doc.number} onChange={(e) => setDoc({ ...doc, number: e.target.value })} /></label>
          <label className="field">Expires<input type="date" value={doc.expiresOn} onChange={(e) => setDoc({ ...doc, expiresOn: e.target.value })} /></label>
          <div style={{ alignSelf: "end" }}><button className="btn small" disabled={act.busy}>Add document</button></div>
        </form>
      )}

      {manage && (
        <>
          <h3>Vehicle and approval</h3>
          <div className="row">
            <select aria-label="Vehicle" value={driver.vehicleId ?? ""} onChange={(e) => e.target.value && go(() => api("PUT", `${base}/drivers/${id}/vehicle`, { vehicleId: e.target.value }))}>
              <option value="">Assign a vehicle…</option>
              {vehicles.map((v) => <option key={v.id} value={v.id}>{v.plate} ({v.status})</option>)}
            </select>
          </div>
          <div className="row">
            {driver.status !== "approved" && <button className="btn primary" disabled={act.busy} onClick={() => go(() => api("POST", `${base}/drivers/${id}/status`, { status: "approved" }))}>Approve driver</button>}
            {driver.status === "approved" && <button className="btn danger" disabled={act.busy} onClick={() => {
              const reason = prompt("Reason for suspending this driver:");
              if (reason) go(() => api("POST", `${base}/drivers/${id}/status`, { status: "suspended", reason }));
            }}>Suspend driver</button>}
          </div>
        </>
      )}
    </div>
  );
}
