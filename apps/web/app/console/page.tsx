"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { Badge, ErrorBox, SignOut, Spinner, TopBar } from "@/components/ui";
import { Ctx, Tenant, can } from "@/components/console/ctx";
import Overview from "@/components/console/Overview";
import Setup from "@/components/console/Setup";
import Drivers from "@/components/console/Drivers";
import Vehicles from "@/components/console/Vehicles";
import Bookings from "@/components/console/Bookings";
import Dispatch from "@/components/console/Dispatch";
import Team from "@/components/console/Team";

const TABS = [
  { id: "overview", label: "Overview", perm: "tenant.read" },
  { id: "dispatch", label: "Dispatch", perm: "dispatch.manage" },
  { id: "bookings", label: "Bookings", perm: "booking.read" },
  { id: "drivers", label: "Drivers", perm: "fleet.read" },
  { id: "vehicles", label: "Vehicles", perm: "fleet.read" },
  { id: "setup", label: "Setup", perm: "tenant.read" },
  { id: "team", label: "Team", perm: "staff.read" },
] as const;

function Console() {
  const t = useSearchParams().get("t") ?? "";
  const [ctx, setCtx] = useState<Ctx | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [tab, setTab] = useState<string>(() => (typeof window !== "undefined" && window.location.hash.slice(1)) || "overview");

  const reload = useCallback(async () => {
    try {
      const r = await api<{ tenant: Tenant; myRole: string }>("GET", `/v1/tenants/${t}`);
      setCtx({ tenantId: t, tenant: r.tenant, role: r.myRole, reload });
    } catch (e) {
      setError(e);
    }
  }, [t]);
  useEffect(() => { if (t) reload(); }, [t, reload]);
  useEffect(() => { history.replaceState(null, "", `#${tab}`); }, [tab]);

  if (!t) return <div className="alert error">No business selected. <Link href="/home/">Choose one</Link>.</div>;
  if (error) return <ErrorBox error={error} />;
  if (!ctx) return <Spinner />;

  const tabs = TABS.filter((x) => can(ctx.role, x.perm));
  const active = tabs.some((x) => x.id === tab) ? tab : "overview";
  return (
    <>
      <div className="row" style={{ marginBottom: 8 }}>
        <h1 style={{ margin: 0 }}>{ctx.tenant.name}</h1>
        <Badge value={ctx.tenant.status} />
        <span className="muted small">You are {ctx.role}</span>
      </div>
      {ctx.tenant.status === "onboarding" && active !== "setup" && (
        <div className="alert warn" style={{ marginBottom: 12 }}>
          This business isn't live yet. <button className="linkbtn" onClick={() => setTab("setup")}>Finish setup</button> to start taking bookings.
        </div>
      )}
      {(ctx.tenant.status === "suspended" || ctx.tenant.status === "closed") && (
        <div className="alert error" style={{ marginBottom: 12 }}>This business is {ctx.tenant.status}. You can view data but not make changes.</div>
      )}
      <nav className="tabs" role="tablist">
        {tabs.map((x) => (
          <button key={x.id} role="tab" className="tab" aria-selected={active === x.id} onClick={() => setTab(x.id)}>{x.label}</button>
        ))}
      </nav>
      {active === "overview" && <Overview ctx={ctx} />}
      {active === "setup" && <Setup ctx={ctx} />}
      {active === "drivers" && <Drivers ctx={ctx} />}
      {active === "vehicles" && <Vehicles ctx={ctx} />}
      {active === "bookings" && <Bookings ctx={ctx} />}
      {active === "dispatch" && <Dispatch ctx={ctx} />}
      {active === "team" && <Team ctx={ctx} />}
    </>
  );
}

export default function Page() {
  return (
    <>
      <TopBar title="Console"><Link className="btn small" href="/home/">Switch</Link><SignOut /></TopBar>
      <main className="page">
        <Suspense><Console /></Suspense>
      </main>
    </>
  );
}
