"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { ErrorBox, Spinner, useAction } from "@/components/ui";
import FilePick from "@/components/FilePick";
import { Ctx, can } from "./ctx";

export type Brand = { name: string; color: string; onColor: string; logoUrl: string | null; tagline: string | null; supportPhone: string | null; supportEmail: string | null; websiteUrl: string | null; termsUrl: string | null; privacyUrl: string | null; language: string; currency: string };

const LINKS: [string, string, string][] = [["supportPhone", "Support phone", "tel"], ["supportEmail", "Support email", "email"], ["websiteUrl", "Website", "url"], ["termsUrl", "Terms link", "url"], ["privacyUrl", "Privacy policy link", "url"]];

/** Runtime branding (T-007): what passengers see on the booking page and in the apps. */
export default function Branding({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const editable = can(ctx.role, "tenant.manage");
  const [brand, setBrand] = useState<Brand | null>(null);
  const [f, setF] = useState<Record<string, string>>({});
  const [logo, setLogo] = useState<{ id: string; fileName: string } | null>(null);
  const act = useAction();
  const load = useCallback(() => api<{ brand: Brand }>("GET", `${base}/branding`).then((r) => {
    setBrand(r.brand);
    setF({ color: r.brand.color, tagline: r.brand.tagline ?? "", supportPhone: r.brand.supportPhone ?? "", supportEmail: r.brand.supportEmail ?? "", websiteUrl: r.brand.websiteUrl ?? "", termsUrl: r.brand.termsUrl ?? "", privacyUrl: r.brand.privacyUrl ?? "" });
  }).catch(act.setError), [base, act.setError]);
  useEffect(() => { load(); }, [load]);
  if (!brand) return <div className="card"><Spinner /></div>;

  const preview = { background: f.color || brand.color, color: brand.onColor };
  return (
    <div className="card stack">
      <h2>Brand</h2>
      <p className="muted small">Shown on your booking page (<span className="mono">/book/?b={ctx.tenant.slug}</span>) and in your passenger and driver apps. Store listings and app icons are set up separately when your branded apps are built.</p>
      <div className="row" style={{ ...preview, padding: 12, borderRadius: 10 }}>
        {brand.logoUrl && <img src={`${brand.logoUrl}?v=${Date.now()}`} alt="" style={{ height: 32, borderRadius: 6, background: "#fff" }} />}
        <strong>{brand.name}</strong>{brand.tagline && <span style={{ opacity: 0.85 }}>— {f.tagline || brand.tagline}</span>}
      </div>
      {editable && (
        <form className="stack" onSubmit={(e) => {
          e.preventDefault();
          act.run(async () => { await api("PATCH", `${base}/branding`, { ...f, ...(logo && { logoFileId: logo.id }) }); setLogo(null); await load(); });
        }}>
          <div className="fields">
            <label className="field">Brand colour
              <div className="row"><input type="color" value={f.color} onChange={(e) => setF({ ...f, color: e.target.value })} /><input className="mono" value={f.color} pattern="#[0-9a-fA-F]{6}" onChange={(e) => setF({ ...f, color: e.target.value })} style={{ maxWidth: 120 }} /></div>
            </label>
            <FilePick key={logo?.id ?? "empty"} base={base} value={logo} onChange={setLogo} label="Logo (PNG/JPEG/WebP)" />
            <label className="field" style={{ gridColumn: "1/-1" }}>Tagline<input maxLength={160} value={f.tagline} onChange={(e) => setF({ ...f, tagline: e.target.value })} placeholder="e.g. Kandy's friendliest cabs, 24/7" /></label>
            {LINKS.map(([k, label, type]) => <label className="field" key={k}>{label}<input type={type} value={f[k]} onChange={(e) => setF({ ...f, [k]: e.target.value })} placeholder={type === "url" ? "https://" : ""} /></label>)}
          </div>
          <ErrorBox error={act.error} />
          <div className="row">
            <button className="btn primary" disabled={act.busy}>Save brand</button>
            {brand.logoUrl && <button type="button" className="btn" disabled={act.busy} onClick={() => act.run(async () => { await api("PATCH", `${base}/branding`, { logoFileId: "" }); await load(); })}>Remove logo</button>}
          </div>
        </form>
      )}
    </div>
  );
}
