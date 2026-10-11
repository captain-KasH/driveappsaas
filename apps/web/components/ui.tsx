"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { api, ApiError, Me, token } from "@/lib/api";

export function TopBar({ title, children }: { title?: string; children?: React.ReactNode }) {
  return (
    <header className="topbar">
      <Link href="/" className="logo">
        <span className="logo-mark" aria-hidden>T</span>
        Taxios
      </Link>
      {title && <span className="muted small">/ {title}</span>}
      <span className="spacer" />
      {children}
    </header>
  );
}

export function SignOut() {
  const router = useRouter();
  return (
    <button
      className="btn small"
      onClick={async () => {
        await api("POST", "/v1/auth/logout").catch(() => {});
        token.clear();
        router.replace("/login/");
      }}
    >
      Sign out
    </button>
  );
}

export function ErrorBox({ error }: { error: unknown }) {
  if (!error) return null;
  const msg = error instanceof Error ? error.message : String(error);
  return <div className="alert error" role="alert">{msg}</div>;
}

export function Spinner({ label = "Loading" }: { label?: string }) {
  return (
    <div className="row muted" role="status">
      <span className="spinner" /> {label}…
    </div>
  );
}

const tone: Record<string, string> = {
  active: "ok", approved: "ok", completed: "ok", available: "ok", accepted: "ok", paid_cash: "ok",
  onboarding: "warn", pending_verification: "warn", pending: "warn", searching: "warn", scheduled: "warn", busy: "warn", assigned: "warn",
  en_route: "warn", arrived: "warn", in_progress: "warn", driver_assigned: "warn",
  suspended: "bad", closed: "bad", cancelled: "bad", expired: "bad", rejected: "bad", aborted: "bad", offline: "",
};

export function Badge({ value }: { value?: string | null }) {
  if (!value) return null;
  return <span className={`badge ${tone[value] ?? ""}`}>{value.replaceAll("_", " ")}</span>;
}

/** Loads /v1/me; sends anonymous visitors to sign in. */
export function useMe() {
  const router = useRouter();
  const [me, setMe] = useState<Me | null>(null);
  const [error, setError] = useState<unknown>(null);
  useEffect(() => {
    if (!token.get()) {
      router.replace(`/login/?next=${encodeURIComponent(location.pathname + location.search)}`);
      return;
    }
    api<Me>("GET", "/v1/me")
      .then(setMe)
      .catch((e) => {
        if (e instanceof ApiError && e.status === 401) router.replace("/login/");
        else setError(e);
      });
  }, [router]);
  return { me, error };
}

/** Runs an async action with busy/error state, returning a stable trigger. */
export function useAction() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const run = useCallback(async <T,>(fn: () => Promise<T>): Promise<T | undefined> => {
    setBusy(true);
    setError(null);
    try {
      return await fn();
    } catch (e) {
      setError(e);
      return undefined;
    } finally {
      setBusy(false);
    }
  }, []);
  return { busy, error, run, setError };
}

/** Polls fn every `ms` while mounted (stops when the tab is hidden). */
export function usePoll(fn: () => void, ms: number) {
  useEffect(() => {
    fn();
    const id = setInterval(() => document.visibilityState === "visible" && fn(), ms);
    return () => clearInterval(id);
  }, [fn, ms]);
}
