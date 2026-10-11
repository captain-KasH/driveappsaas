"use client";

import { useCallback, useEffect, useState } from "react";
import { ago, api, when } from "@/lib/api";
import { ErrorBox, useAction } from "@/components/ui";

type Session = { id: string; client: string; userAgent: string | null; createdAt: string; lastSeenAt: string | null; expiresAt: string; current: boolean };

/** Where you're signed in (T-112). */
export default function Sessions() {
  const [list, setList] = useState<Session[] | null>(null);
  const act = useAction();
  const load = useCallback(() => api<{ sessions: Session[] }>("GET", "/v1/auth/sessions").then((r) => setList(r.sessions)).catch(act.setError), [act.setError]);
  useEffect(() => { load(); }, [load]);
  const short = (ua: string | null) => (ua ?? "unknown device").replace(/Mozilla\/5\.0 \(([^)]*)\).*/, "$1").slice(0, 60);
  return (
    <details className="card">
      <summary><strong>Where you're signed in</strong> <span className="muted small">{list ? `${list.length} session${list.length === 1 ? "" : "s"}` : ""}</span></summary>
      <div className="stack" style={{ marginTop: 12 }}>
        <ErrorBox error={act.error} />
        {list?.map((s) => (
          <div key={s.id} className="row small">
            <span><strong>{s.client === "mobile" ? "Mobile app" : "Browser"}</strong>{s.current && <span className="badge ok" style={{ marginLeft: 6 }}>this one</span>}</span>
            <span className="muted">{short(s.userAgent)} · active {ago(s.lastSeenAt ?? s.createdAt)} · since {when(s.createdAt)}</span>
            {!s.current && <button className="btn small right" disabled={act.busy} onClick={() => act.run(async () => { await api("DELETE", `/v1/auth/sessions/${s.id}`); await load(); })}>Sign out</button>}
          </div>
        ))}
        {list && list.length > 1 && (
          <div><button className="btn small danger" disabled={act.busy} onClick={() => act.run(async () => { await api("POST", "/v1/auth/logout-all"); await load(); })}>Sign out everywhere else</button></div>
        )}
      </div>
    </details>
  );
}
