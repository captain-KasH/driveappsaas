"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { api } from "@/lib/api";
import { BUSINESS_TYPES } from "@/lib/business";
import { ErrorBox, SignOut, Spinner, TopBar, useAction, useMe } from "@/components/ui";

const slugify = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 63);

/** Step 1 of self-service onboarding: a verified account opens a draft business (D-021). */
export default function Start() {
  const router = useRouter();
  const { me, error: meError } = useMe();
  const [f, setF] = useState({ name: "", slug: "", businessType: "taxi" });
  const [slugTouched, setSlugTouched] = useState(false);
  const create = useAction();
  const resend = useAction();
  const [resent, setResent] = useState(false);

  if (meError) return <><TopBar /><main className="page narrow"><ErrorBox error={meError} /></main></>;
  if (!me) return <><TopBar /><main className="page narrow"><Spinner /></main></>;

  return (
    <>
      <TopBar title="New business"><Link className="btn small" href="/home/">Home</Link><SignOut /></TopBar>
      <main className="page narrow">
        {!me.user.emailVerified ? (
          <div className="card stack">
            <h1>Check your email</h1>
            <p>We sent a verification link to <strong>{me.user.email}</strong>. Open it, then come back here to create your business.</p>
            <ErrorBox error={resend.error} />
            {resent ? <div className="alert ok">Sent again. Check your spam folder too.</div> : (
              <div><button className="btn" disabled={resend.busy} onClick={() => resend.run(async () => { await api("POST", "/v1/auth/resend-verification"); setResent(true); })}>Resend the link</button></div>
            )}
            <p className="muted small">Wrong address? Sign out and create a new account.</p>
          </div>
        ) : (
          <form className="card stack" onSubmit={(e) => {
            e.preventDefault();
            create.run(async () => {
              const r = await api<{ tenant: { id: string } }>("POST", "/v1/businesses", f);
              router.replace(`/console/?t=${r.tenant.id}#setup`);
            });
          }}>
            <h1>Create your transport business</h1>
            <p className="muted small">You can change the name later. The booking link is permanent once passengers start using it.</p>
            <label className="field">Business name
              <input required maxLength={200} value={f.name} placeholder="Kandy Cabs"
                onChange={(e) => setF({ ...f, name: e.target.value, slug: slugTouched ? f.slug : slugify(e.target.value) })} />
            </label>
            <label className="field">Booking link
              <div className="row" style={{ gap: 4, flexWrap: "nowrap" }}>
                <span className="muted small mono">taxios.packline.lk/book/?b=</span>
                <input required pattern="[a-z0-9][a-z0-9-]{1,61}[a-z0-9]" title="3–63 lowercase letters, digits or dashes" value={f.slug}
                  onChange={(e) => { setSlugTouched(true); setF({ ...f, slug: e.target.value.toLowerCase() }); }} />
              </div>
            </label>
            <label className="field">Type of business
              <select value={f.businessType} onChange={(e) => setF({ ...f, businessType: e.target.value })}>
                {BUSINESS_TYPES.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
              </select>
            </label>
            <ErrorBox error={create.error} />
            <button className="btn primary block" disabled={create.busy}>{create.busy ? "Creating…" : "Create business"}</button>
            <p className="muted small">Next: business details, service area, vehicles and fares. Taxios reviews the details before your booking page goes public.</p>
          </form>
        )}
      </main>
    </>
  );
}
