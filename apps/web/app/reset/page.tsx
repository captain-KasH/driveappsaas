"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { api, token } from "@/lib/api";
import { ErrorBox, TopBar, useAction } from "@/components/ui";

function Reset() {
  const router = useRouter();
  const t = useSearchParams().get("token") ?? "";
  const [password, setPassword] = useState("");
  const { busy, error, run } = useAction();
  if (!t) return <div className="alert error">This link is missing its token. <Link href="/forgot/">Request a new one</Link>.</div>;
  return (
    <form className="card stack" onSubmit={(e) => {
      e.preventDefault();
      run(async () => {
        const r = await api<{ token: string }>("POST", "/v1/auth/reset-password", { token: t, password });
        token.set(r.token); // every other session was signed out by the server
        router.replace("/home/");
      });
    }}>
      <h1>Choose a new password</h1>
      <label className="field">New password (10+ characters)<input type="password" required minLength={10} maxLength={72} autoComplete="new-password" value={password} onChange={(e) => setPassword(e.target.value)} /></label>
      <ErrorBox error={error} />
      <button className="btn primary block" disabled={busy}>{busy ? "Saving…" : "Save and sign in"}</button>
      <p className="muted small">All your other sessions will be signed out.</p>
    </form>
  );
}

export default function Page() {
  return (
    <>
      <TopBar />
      <main className="page narrow"><Suspense><Reset /></Suspense></main>
    </>
  );
}
