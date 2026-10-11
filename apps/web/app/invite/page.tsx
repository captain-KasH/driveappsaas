"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { api, token } from "@/lib/api";
import { ErrorBox, TopBar, useAction } from "@/components/ui";

function Accept() {
  const router = useRouter();
  const invite = useSearchParams().get("token") ?? "";
  const [fullName, setFullName] = useState("");
  const [password, setPassword] = useState("");
  const { busy, error, run } = useAction();

  if (!invite) return <div className="alert error">This invitation link is incomplete. Ask whoever invited you to send it again.</div>;

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    run(async () => {
      const r = await api<{ token: string; tenantId: string; role: string }>("POST", "/v1/auth/accept-invite", { token: invite, fullName, password });
      token.set(r.token);
      router.replace(r.role === "driver" ? `/driver/?t=${r.tenantId}` : `/console/?t=${r.tenantId}`);
    });
  };

  return (
    <form className="card stack" onSubmit={submit}>
      <h1>Accept your invitation</h1>
      <p className="muted small">Set up your Taxios account. If you already have one with this email, enter its password instead — your name won't change.</p>
      <label className="field">Full name<input autoComplete="name" value={fullName} onChange={(e) => setFullName(e.target.value)} /></label>
      <label className="field">Password (10–72 characters)<input type="password" autoComplete="new-password" required minLength={10} maxLength={72} value={password} onChange={(e) => setPassword(e.target.value)} /></label>
      <ErrorBox error={error} />
      <button className="btn primary block" disabled={busy}>{busy ? "Setting up…" : "Accept and continue"}</button>
    </form>
  );
}

export default function Page() {
  return (
    <>
      <TopBar />
      <main className="page narrow">
        <Suspense><Accept /></Suspense>
      </main>
    </>
  );
}
