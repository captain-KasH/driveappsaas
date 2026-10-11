"use client";

import Link from "next/link";
import { Badge, ErrorBox, SignOut, Spinner, TopBar, useMe } from "@/components/ui";

const roleLabel: Record<string, string> = {
  owner: "Owner", admin: "Admin", dispatcher: "Dispatcher", finance: "Finance", viewer: "Viewer", driver: "Driver", passenger: "Passenger",
};

export default function Home() {
  const { me, error } = useMe();
  return (
    <>
      <TopBar><SignOut /></TopBar>
      <main className="page">
        <ErrorBox error={error} />
        {!me ? <Spinner /> : (
          <div className="stack">
            <h1>Hi {me.user.fullName.split(" ")[0]}</h1>
            {me.user.platformRole && (
              <div className="card row">
                <div>
                  <h3>Platform administration</h3>
                  <p className="muted small">Create and manage transport businesses on Taxios.</p>
                </div>
                <Link className="btn primary right" href="/admin/">Open admin</Link>
              </div>
            )}
            {me.memberships.length === 0 && !me.user.platformRole && (
              <div className="card empty">You're not part of any business yet. Ask an owner to invite you.</div>
            )}
            {me.memberships.map((m) => (
              <div className="card row" key={m.tenantId}>
                <div>
                  <h3>{m.tenantName}</h3>
                  <div className="row small"><span className="muted">{roleLabel[m.role] ?? m.role}</span><Badge value={m.tenantStatus} /></div>
                </div>
                <Link className="btn primary right" href={m.role === "driver" ? `/driver/?t=${m.tenantId}` : m.role === "passenger" ? `/book/?b=${m.tenantSlug}` : `/console/?t=${m.tenantId}`}>
                  {m.role === "driver" ? "Driver app" : m.role === "passenger" ? "Book a ride" : "Open console"}
                </Link>
              </div>
            ))}
          </div>
        )}
      </main>
    </>
  );
}
