"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { api, token } from "@/lib/api";
import { ErrorBox, TopBar, useAction } from "@/components/ui";

function Login() {
  const router = useRouter();
  const next = useSearchParams().get("next");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const { busy, error, run } = useAction();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    run(async () => {
      const r = await api<{ token: string }>("POST", "/v1/auth/login", { email, password });
      token.set(r.token);
      // Only same-site paths are honoured as a redirect target.
      router.replace(next && next.startsWith("/") && !next.startsWith("//") ? next : "/home/");
    });
  };

  return (
    <form className="card stack" onSubmit={submit}>
      <h1>Sign in</h1>
      <label className="field">Email<input type="email" autoComplete="username" required value={email} onChange={(e) => setEmail(e.target.value)} /></label>
      <label className="field">Password<input type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} /></label>
      <ErrorBox error={error} />
      <button className="btn primary block" disabled={busy}>{busy ? "Signing in…" : "Sign in"}</button>
      <p className="muted small">New staff member or driver? Open the invitation link you were sent to set your password.</p>
    </form>
  );
}

export default function Page() {
  return (
    <>
      <TopBar />
      <main className="page narrow">
        <Suspense><Login /></Suspense>
      </main>
    </>
  );
}
