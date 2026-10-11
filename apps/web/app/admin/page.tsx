"use client";

import { useCallback, useEffect, useState } from "react";
import { api, when } from "@/lib/api";
import { Badge, ErrorBox, SignOut, Spinner, TopBar, useAction, useMe } from "@/components/ui";
import { CopyLink } from "@/components/CopyLink";

type Tenant = { id: string; slug: string; name: string; businessType: string; countryCode: string; status: string; createdAt: string };

const TYPES = ["taxi", "three_wheeler", "independent_driver", "airport_transfer", "chauffeur", "tourism", "hospitality", "corporate", "shuttle", "startup", "enterprise_fleet"];

export default function Admin() {
  const { me, error: meError } = useMe();
  const [tenants, setTenants] = useState<Tenant[] | null>(null);
  const [invite, setInvite] = useState<{ name: string; email: string; link: string } | null>(null);
  const [form, setForm] = useState({ name: "", slug: "", businessType: "taxi", ownerEmail: "" });
  const create = useAction();
  const status = useAction();

  const load = useCallback(() => api<{ tenants: Tenant[] }>("GET", "/v1/platform/tenants").then((r) => setTenants(r.tenants)).catch(status.setError), [status.setError]);
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

  return (
    <>
      <TopBar title="Platform admin"><SignOut /></TopBar>
      <main className="page stack">
        <ErrorBox error={meError} />
        {canSuper && (
          <form className="card stack" onSubmit={submit}>
            <h2>New transport business</h2>
            <div className="fields">
              <label className="field">Business name<input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label>
              <label className="field">Booking link slug
                <input required pattern="[a-z0-9][a-z0-9-]{1,61}[a-z0-9]" title="3–63 lowercase letters, digits or dashes" placeholder="colombo-cabs"
                  value={form.slug} onChange={(e) => setForm({ ...form, slug: e.target.value.toLowerCase() })} />
              </label>
              <label className="field">Type
                <select value={form.businessType} onChange={(e) => setForm({ ...form, businessType: e.target.value })}>
                  {TYPES.map((t) => <option key={t} value={t}>{t.replaceAll("_", " ")}</option>)}
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
        )}
        <div className="card">
          <h2>Businesses</h2>
          <ErrorBox error={status.error} />
          {!tenants ? <Spinner /> : tenants.length === 0 ? <div className="empty">No businesses yet.</div> : (
            <table className="list">
              <thead><tr><th>Name</th><th>Type</th><th>Status</th><th>Created</th><th></th></tr></thead>
              <tbody>
                {tenants.map((t) => (
                  <tr key={t.id}>
                    <td data-label="Name"><strong>{t.name}</strong><div className="muted small mono">/book/?b={t.slug}</div></td>
                    <td data-label="Type">{t.businessType.replaceAll("_", " ")}</td>
                    <td data-label="Status"><Badge value={t.status} /></td>
                    <td data-label="Created">{when(t.createdAt)}</td>
                    <td data-label="">
                      {canSuper && (
                        <div className="row" style={{ gap: 6 }}>
                          {t.status === "suspended" && <button className="btn small" onClick={() => change(t, "active")}>Reactivate</button>}
                          {(t.status === "active" || t.status === "onboarding") && <button className="btn small" onClick={() => change(t, "suspended")}>Suspend</button>}
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
