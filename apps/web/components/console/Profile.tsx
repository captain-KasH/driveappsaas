"use client";

import { useState } from "react";
import { api } from "@/lib/api";
import { businessTypeLabel } from "@/lib/business";
import { ErrorBox, useAction } from "@/components/ui";
import { Ctx, can } from "./ctx";

const FIELDS: [string, string, string?][] = [
  ["name", "Display name"], ["legalName", "Legal / registered name"], ["registrationNo", "Business registration no."], ["taxNo", "Tax (TIN / VAT) no."],
  ["contactEmail", "Contact email", "email"], ["contactPhone", "Contact phone", "tel"], ["billingEmail", "Billing email", "email"], ["address", "Address"],
];

/** Business profile: the details Taxios verifies and that appear on receipts. */
export default function Profile({ ctx, onChange }: { ctx: Ctx; onChange: () => Promise<void> }) {
  const t = ctx.tenant;
  const editable = can(ctx.role, "tenant.manage");
  const initial = () => Object.fromEntries([...FIELDS.map(([k]) => [k, (t as any)[k] ?? ""]), ["timezone", t.timezone], ["currency", t.currency], ["language", t.language]]) as Record<string, string>;
  const [f, setF] = useState<Record<string, string>>(initial);
  const [open, setOpen] = useState(!t.legalName);
  const act = useAction();
  const [saved, setSaved] = useState(false);

  const summary = [t.legalName, t.registrationNo && `Reg. ${t.registrationNo}`, t.contactPhone, t.contactEmail].filter(Boolean).join(" · ");
  return (
    <div className="card stack">
      <div className="row">
        <h2 style={{ margin: 0 }}>Business details</h2>
        <span className="muted small">{businessTypeLabel(t.businessType)} · {t.countryCode} · {t.currency} · {t.timezone}</span>
        {editable && !open && <button className="btn small right" onClick={() => setOpen(true)}>Edit</button>}
      </div>
      {!open && <p className="muted small">{summary || "Not filled in yet."}</p>}
      {open && (
        <form className="stack" onSubmit={(e) => {
          e.preventDefault();
          act.run(async () => {
            await api("PATCH", `/v1/tenants/${ctx.tenantId}`, f);
            setSaved(true); setOpen(false);
            await onChange();
          });
        }}>
          <div className="fields">
            {FIELDS.map(([k, label, type]) => (
              <label className="field" key={k}>{label}<input type={type ?? "text"} value={f[k]} disabled={!editable} onChange={(e) => setF({ ...f, [k]: e.target.value })} /></label>
            ))}
            <label className="field">Language
              <select value={f.language} disabled={!editable} onChange={(e) => setF({ ...f, language: e.target.value })}>
                <option value="en">English</option><option value="si">සිංහල</option><option value="ta">தமிழ்</option>
              </select>
            </label>
            <label className="field">Timezone<input value={f.timezone} disabled={!editable} onChange={(e) => setF({ ...f, timezone: e.target.value })} /></label>
            <label className="field">Currency<input value={f.currency} maxLength={3} disabled={!editable} onChange={(e) => setF({ ...f, currency: e.target.value.toUpperCase() })} /></label>
          </div>
          <ErrorBox error={act.error} />
          {editable && (
            <div className="row">
              <button className="btn primary" disabled={act.busy}>Save details</button>
              <button type="button" className="btn" onClick={() => { setF(initial()); setOpen(false); }}>Cancel</button>
            </div>
          )}
        </form>
      )}
      {saved && !open && <div className="alert ok">Saved.</div>}
    </div>
  );
}
