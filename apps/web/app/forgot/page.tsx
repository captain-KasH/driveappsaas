"use client";

import Link from "next/link";
import { useState } from "react";
import { api } from "@/lib/api";
import { ErrorBox, TopBar, useAction } from "@/components/ui";

export default function Forgot() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const { busy, error, run } = useAction();
  return (
    <>
      <TopBar><Link className="btn small" href="/login/">Sign in</Link></TopBar>
      <main className="page narrow">
        <form className="card stack" onSubmit={(e) => { e.preventDefault(); run(async () => { await api("POST", "/v1/auth/forgot-password", { email }); setSent(true); }); }}>
          <h1>Reset your password</h1>
          {sent ? (
            <div className="alert ok">If an account exists for {email}, a reset link is on its way. It works once and expires in 1 hour.</div>
          ) : (
            <>
              <label className="field">Email<input type="email" required autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} /></label>
              <ErrorBox error={error} />
              <button className="btn primary block" disabled={busy}>{busy ? "Sending…" : "Send reset link"}</button>
            </>
          )}
        </form>
      </main>
    </>
  );
}
