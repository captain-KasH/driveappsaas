"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { api, token } from "@/lib/api";
import { ErrorBox, TopBar, useAction } from "@/components/ui";

export default function Signup() {
  const router = useRouter();
  const [f, setF] = useState({ fullName: "", email: "", password: "" });
  const { busy, error, run } = useAction();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    run(async () => {
      const r = await api<{ token: string }>("POST", "/v1/auth/register", f);
      token.set(r.token);
      router.replace("/start/");
    });
  };

  return (
    <>
      <TopBar><Link className="btn small" href="/login/">Sign in</Link></TopBar>
      <main className="page narrow">
        <form className="card stack" onSubmit={submit}>
          <h1>Create your account</h1>
          <p className="muted small">For transport business owners. Drivers and staff join through an invitation link instead.</p>
          <label className="field">Your name<input required autoComplete="name" value={f.fullName} onChange={(e) => setF({ ...f, fullName: e.target.value })} /></label>
          <label className="field">Email<input type="email" required autoComplete="email" value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></label>
          <label className="field">Password (10+ characters)<input type="password" required minLength={10} maxLength={72} autoComplete="new-password" value={f.password} onChange={(e) => setF({ ...f, password: e.target.value })} /></label>
          <ErrorBox error={error} />
          <button className="btn primary block" disabled={busy}>{busy ? "Creating…" : "Continue"}</button>
          <p className="muted small">Already have an account? <Link href="/login/">Sign in</Link>.</p>
        </form>
      </main>
    </>
  );
}
