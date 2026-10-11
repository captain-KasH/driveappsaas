"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { Badge, ErrorBox, Spinner, useAction } from "@/components/ui";
import { Category, Ctx, Vehicle, can } from "./ctx";

export default function Vehicles({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const manage = can(ctx.role, "fleet.manage");
  const [vehicles, setVehicles] = useState<Vehicle[] | null>(null);
  const [cats, setCats] = useState<Category[]>([]);
  const [f, setF] = useState({ categoryId: "", plate: "", make: "", model: "", colour: "" });
  const act = useAction();
  const load = useCallback(() => api<{ vehicles: Vehicle[] }>("GET", `${base}/vehicles`).then((r) => setVehicles(r.vehicles)).catch(act.setError), [base, act.setError]);
  useEffect(() => { load(); api<{ categories: Category[] }>("GET", `${base}/vehicle-categories`).then((r) => setCats(r.categories)).catch(() => {}); }, [load, base]);
  const cat = (id: string) => cats.find((c) => c.id === id)?.name ?? "";

  const setStatus = (v: Vehicle, status: string) => act.run(async () => { await api("POST", `${base}/vehicles/${v.id}/status`, { status }); load(); });

  return (
    <div className="stack">
      {manage && (
        <form className="card stack" onSubmit={(e) => { e.preventDefault(); act.run(async () => { await api("POST", `${base}/vehicles`, f); setF({ ...f, plate: "", make: "", model: "", colour: "" }); load(); }); }}>
          <h2>Add a vehicle</h2>
          <div className="fields">
            <label className="field">Category<select required value={f.categoryId} onChange={(e) => setF({ ...f, categoryId: e.target.value })}>
              <option value="">Choose…</option>{cats.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label>
            <label className="field">Plate<input required placeholder="CAB-1234" value={f.plate} onChange={(e) => setF({ ...f, plate: e.target.value })} /></label>
            <label className="field">Make<input value={f.make} placeholder="Toyota" onChange={(e) => setF({ ...f, make: e.target.value })} /></label>
            <label className="field">Model<input value={f.model} placeholder="Axio" onChange={(e) => setF({ ...f, model: e.target.value })} /></label>
            <label className="field">Colour<input value={f.colour} placeholder="White" onChange={(e) => setF({ ...f, colour: e.target.value })} /></label>
          </div>
          <div><button className="btn primary" disabled={act.busy}>Add vehicle</button></div>
        </form>
      )}
      <ErrorBox error={act.error} />
      <div className="card">
        <h2>Vehicles</h2>
        {!vehicles ? <Spinner /> : vehicles.length === 0 ? <div className="empty">No vehicles yet.</div> : (
          <table className="list">
            <thead><tr><th>Plate</th><th>Vehicle</th><th>Category</th><th>Status</th><th></th></tr></thead>
            <tbody>
              {vehicles.map((v) => (
                <tr key={v.id}>
                  <td data-label="Plate"><strong className="mono">{v.plate}</strong></td>
                  <td data-label="Vehicle">{[v.colour, v.make, v.model].filter(Boolean).join(" ") || "—"}</td>
                  <td data-label="Category">{cat(v.categoryId)}</td>
                  <td data-label="Status"><Badge value={v.status} /></td>
                  <td>{manage && (v.status === "approved"
                    ? <button className="btn small danger" disabled={act.busy} onClick={() => setStatus(v, "suspended")}>Suspend</button>
                    : <button className="btn small" disabled={act.busy} onClick={() => setStatus(v, "approved")}>Approve</button>)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
