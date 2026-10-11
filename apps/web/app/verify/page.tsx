"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { api, token } from "@/lib/api";
import { ErrorBox, Spinner, TopBar } from "@/components/ui";

function Verify() {
  const t = useSearchParams().get("token") ?? "";
  const [state, setState] = useState<"working" | "done" | "failed">("working");
  const [error, setError] = useState<unknown>(null);
  useEffect(() => {
    if (!t) { setState("failed"); setError(new Error("This link is missing its token.")); return; }
    api("POST", "/v1/auth/verify-email", { token: t }).then(() => setState("done")).catch((e) => { setError(e); setState("failed"); });
  }, [t]);
  const signedIn = typeof window !== "undefined" && !!token.get();
  return (
    <div className="card stack">
      <h1>Email verification</h1>
      {state === "working" && <Spinner label="Checking your link" />}
      {state === "done" && (
        <>
          <div className="alert ok">Your email is verified.</div>
          <Link className="btn primary" href={signedIn ? "/start/" : "/login/?next=%2Fstart%2F"}>{signedIn ? "Set up your business" : "Sign in to continue"}</Link>
        </>
      )}
      {state === "failed" && (
        <>
          <ErrorBox error={error} />
          <p className="muted small">Links work once and expire after 24 hours. <Link href="/home/">Sign in</Link> and request a new one from your home page.</p>
        </>
      )}
    </div>
  );
}

export default function Page() {
  return (
    <>
      <TopBar />
      <main className="page narrow"><Suspense><Verify /></Suspense></main>
    </>
  );
}
