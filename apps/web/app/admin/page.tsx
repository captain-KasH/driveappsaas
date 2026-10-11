"use client";

import { useCallback, useEffect, useState } from "react";
import { api, openFile, when } from "@/lib/api";
import { BUSINESS_TYPES, businessTypeLabel } from "@/lib/business";
import { Badge, ErrorBox, SignOut, Spinner, TopBar, useAction, useMe } from "@/components/ui";
import { CopyLink } from "@/components/CopyLink";
import type { Tenant } from "@/components/console/ctx";

const STATUSES = ["", "pending_verification", "onboarding", "active", "suspended", "closed"];

export default function Admin() {
  const { me, error: meError } = useMe();
  const [tenants, setTenants] = useState<Tenant[] | null>(null);
  const [filter, setFilter] = useState({ status: "", q: "" });
  const [open, setOpen] = useState<string | null>(null);
  const [invite, setInvite] = useState<{ name: string; email: string; link: string } | null>(null);
  const [form, setForm] = useState({ name: "", slug: "", businessType: "taxi", ownerEmail: "" });
  const create = useAction();
  const status = useAction();

  const load = useCallback(() => {
    const qs = new URLSearchParams(Object.entries(filter).filter(([, v]) => v)).toString();
    return api<{ tenants: Tenant[] }>("GET", `/v1/platform/tenants${qs ? "?" + qs : ""}`).then((r) => setTenants(r.tenants)).catch(status.setError);
  }, [filter, status.setError]);
  useEffect(() => { if (me?.user.platformRole) load(); }, [me, load]);

  if (me && !me.user.platformRole) return <><TopBar title="Admin"><SignOut /></TopBar><main className="page"><div className="alert error">This area is for Taxios platform staff.</div></main></>;

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    create.run(async () => {
      const r = await api<{ tenant: Tenant; ownerInvitation: { email: string; token: string } }>("POST", "/v1/platform/tenants", form);
      setInvite({ name: r.tenant.name, email: r.ownerInvitation.email, link: `${location.origin}/invite/?token=${r.ownerInvitation.token}` });
      setForm({ name: "", slug: "", businessType: "taxi", ownerEmail: "" });
      load();
    });
  };

  const change = (t: Tenant, to: string) => {
    const reason = prompt(`Reason for moving ${t.name} to "${to}" (recorded in the audit log):`);
    if (!reason) return;
    status.run(async () => { await api("POST", `/v1/platform/tenants/${t.id}/status`, { status: to, reason }); load(); });
  };

  const canSuper = me?.user.platformRole === "super_admin";
  const pending = tenants?.filter((t) => t.status === "pending_verification") ?? [];

  return (
    <>
      <TopBar title="Platform admin"><SignOut /></TopBar>
      <main className="page stack">
        <ErrorBox error={meError} />
        {canSuper && pending.length > 0 && (
          <div className="card stack">
            <h2>Verification queue <span className="badge warn">{pending.length}</span></h2>
            <p className="muted small">Self-registered businesses waiting for approval. Check the registration number against the certificate before approving.</p>
            {pending.map((t) => <VerifyCard key={t.id} t={t} onDone={load} />)}
          </div>
        )}
        {canSuper && (
          <details className="card">
            <summary><strong>Create a business directly</strong> <span className="muted small">(trusted — skips verification)</span></summary>
            <form className="stack" style={{ marginTop: 12 }} onSubmit={submit}>
              <div className="fields">
                <label className="field">Business name<input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label>
                <label className="field">Booking link slug
                  <input required pattern="[a-z0-9][a-z0-9-]{1,61}[a-z0-9]" title="3–63 lowercase letters, digits or dashes" placeholder="colombo-cabs"
                    value={form.slug} onChange={(e) => setForm({ ...form, slug: e.target.value.toLowerCase() })} />
                </label>
                <label className="field">Type
                  <select value={form.businessType} onChange={(e) => setForm({ ...form, businessType: e.target.value })}>
                    {BUSINESS_TYPES.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                  </select>
                </label>
                <label className="field">Owner email<input type="email" required value={form.ownerEmail} onChange={(e) => setForm({ ...form, ownerEmail: e.target.value })} /></label>
              </div>
              <ErrorBox error={create.error} />
              <div><button className="btn primary" disabled={create.busy}>Create business</button></div>
              {invite && (
                <div className="alert ok stack">
                  <strong>{invite.name} created.</strong>
                  <span>Send this one-time link to {invite.email} so they can set a password (expires in 7 days). It won't be shown again.</span>
                  <CopyLink value={invite.link} />
                </div>
              )}
            </form>
          </details>
        )}
        {canSuper && <ManualVerify />}
        <div className="card stack">
          <div className="row">
            <h2 style={{ margin: 0 }}>Businesses</h2>
            <input className="right" placeholder="Search name, slug, legal name" value={filter.q} onChange={(e) => setFilter({ ...filter, q: e.target.value })} style={{ maxWidth: 260 }} />
            <select aria-label="Status" value={filter.status} onChange={(e) => setFilter({ ...filter, status: e.target.value })}>
              {STATUSES.map((s) => <option key={s} value={s}>{s ? s.replaceAll("_", " ") : "All statuses"}</option>)}
            </select>
          </div>
          <ErrorBox error={status.error} />
          {!tenants ? <Spinner /> : tenants.length === 0 ? <div className="empty">No businesses match.</div> : (
            <table className="list">
              <thead><tr><th>Name</th><th>Type</th><th>Status</th><th>Verified</th><th>Created</th><th></th></tr></thead>
              <tbody>
                {tenants.map((t) => (
                  <tr key={t.id}>
                    <td data-label="Name">
                      <button className="linkbtn" onClick={() => setOpen(open === t.id ? null : t.id)}><strong>{t.name}</strong></button>
                      <div className="muted small mono">/book/?b={t.slug}</div>
                      {open === t.id && <TenantDetail t={t} />}
                    </td>
                    <td data-label="Type">{businessTypeLabel(t.businessType)}</td>
                    <td data-label="Status"><Badge value={t.status} />{t.statusReason && <div className="muted small">{t.statusReason}</div>}</td>
                    <td data-label="Verified">{t.verifiedAt ? "✓" : <span className="muted">no</span>}</td>
                    <td data-label="Created">{when(t.createdAt)}</td>
                    <td data-label="">
                      {canSuper && (
                        <div className="row" style={{ gap: 6 }}>
                          {t.status === "suspended" && <button className="btn small" onClick={() => change(t, "active")}>Reactivate</button>}
                          {["active", "onboarding", "pending_verification"].includes(t.status) && <button className="btn small" onClick={() => change(t, "suspended")}>Suspend</button>}
                          {t.status !== "closed" && <button className="btn small danger" onClick={() => change(t, "closed")}>Close</button>}
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>
    </>
  );
}

function TenantDetail({ t }: { t: Tenant }) {
  const rows: [string, string | null | undefined][] = [
    ["Legal name", t.legalName], ["Registration no.", t.registrationNo], ["Tax no.", t.taxNo], ["Contact", [t.contactEmail, t.contactPhone].filter(Boolean).join(" · ")],
    ["Billing email", t.billingEmail], ["Address", t.address], ["Locale", `${t.countryCode} · ${t.currency} · ${t.timezone} · ${t.language}`],
    ["Submitted", t.submittedAt && when(t.submittedAt)], ["Verified", t.verifiedAt && when(t.verifiedAt)],
  ];
  return (
    <dl className="small" style={{ margin: "8px 0 0", display: "grid", gridTemplateColumns: "auto 1fr", gap: "2px 12px" }}>
      {rows.map(([k, v]) => <div key={k} style={{ display: "contents" }}><dt className="muted">{k}</dt><dd style={{ margin: 0 }}>{v || "—"}</dd></div>)}
    </dl>
  );
}

type Detail = { documents: { id: string; docType: string; fileId: string; fileName: string; note: string | null }[]; owners: { email: string; fullName: string; emailVerified: boolean }[] };

function VerifyCard({ t, onDone }: { t: Tenant; onDone: () => void }) {
  const [note, setNote] = useState("");
  const [detail, setDetail] = useState<Detail | null>(null);
  const act = useAction();
  useEffect(() => { api<Detail>("GET", `/v1/platform/tenants/${t.id}`).then(setDetail).catch(act.setError); }, [t.id, act.setError]);
  const decide = (decision: "approve" | "reject") => act.run(async () => {
    await api("POST", `/v1/platform/tenants/${t.id}/verification`, { decision, note });
    onDone();
  });
  return (
    <div className="stack" style={{ borderTop: "1px solid var(--line)", paddingTop: 10 }}>
      <div className="row"><strong>{t.name}</strong><span className="muted small">{businessTypeLabel(t.businessType)} · submitted {when(t.submittedAt)}</span></div>
      <TenantDetail t={t} />
      {detail && (
        <div className="small stack" style={{ gap: 4 }}>
          <div><span className="muted">Owner:</span> {detail.owners.map((o) => `${o.fullName} <${o.email}>${o.emailVerified ? "" : " (email unverified)"}`).join(", ") || "—"}</div>
          <div><span className="muted">Documents:</span> {detail.documents.length === 0 ? <span className="badge bad">none uploaded</span> : detail.documents.map((d) => (
            <button key={d.id} className="linkbtn" style={{ marginRight: 8 }} onClick={() => openFile(`/v1/platform/files/${d.fileId}`).catch(act.setError)}>{d.docType.replaceAll("_", " ")}: {d.fileName}</button>
          ))}</div>
        </div>
      )}
      <label className="field">Note to the owner (required to reject)<input value={note} onChange={(e) => setNote(e.target.value)} placeholder="e.g. registration number doesn't match the certificate" /></label>
      <ErrorBox error={act.error} />
      <div className="row">
        <button className="btn primary" disabled={act.busy} onClick={() => decide("approve")}>Approve and go live</button>
        <button className="btn danger" disabled={act.busy || note.trim().length < 3} onClick={() => decide("reject")}>Send back</button>
      </div>
    </div>
  );
}

/** Until an email provider is connected (K-019), staff can confirm an owner's address by hand. */
function ManualVerify() {
  const [email, setEmail] = useState("");
  const [done, setDone] = useState<string | null>(null);
  const act = useAction();
  return (
    <details className="card">
      <summary><strong>Verify an email address manually</strong> <span className="muted small">(no email provider connected yet — audited)</span></summary>
      <form className="row" style={{ marginTop: 12 }} onSubmit={(e) => { e.preventDefault(); act.run(async () => { await api("POST", "/v1/platform/users/verify-email", { email }); setDone(email); setEmail(""); }); }}>
        <input type="email" required placeholder="owner@example.lk" value={email} onChange={(e) => setEmail(e.target.value)} />
        <button className="btn primary" disabled={act.busy}>Mark verified</button>
      </form>
      <ErrorBox error={act.error} />
      {done && <div className="alert ok" style={{ marginTop: 8 }}>{done} marked as verified.</div>}
    </details>
  );
}
