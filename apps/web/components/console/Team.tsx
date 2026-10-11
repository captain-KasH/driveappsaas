"use client";

import { useCallback, useEffect, useState } from "react";
import { api, when } from "@/lib/api";
import { ErrorBox, Spinner, useAction } from "@/components/ui";
import { CopyLink } from "@/components/CopyLink";
import { Ctx, can } from "./ctx";

type Member = { userId: string; email: string; fullName: string; role: string; since: string };

export default function Team({ ctx }: { ctx: Ctx }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const [members, setMembers] = useState<Member[] | null>(null);
  const [f, setF] = useState({ email: "", role: "dispatcher" });
  const [link, setLink] = useState("");
  const act = useAction();
  const load = useCallback(() => api<{ members: Member[] }>("GET", `${base}/members`).then((r) => setMembers(r.members)).catch(act.setError), [base, act.setError]);
  useEffect(() => { load(); }, [load]);
  const roles = ctx.role === "owner" ? ["owner", "admin", "dispatcher", "finance", "viewer"] : ["admin", "dispatcher", "finance", "viewer"];

  return (
    <div className="stack">
      {can(ctx.role, "staff.manage") && (
        <form className="card stack" onSubmit={(e) => { e.preventDefault(); act.run(async () => {
          const r = await api<{ invitation: { token: string } }>("POST", `${base}/invitations`, f);
          setLink(`${location.origin}/invite/?token=${r.invitation.token}`); setF({ ...f, email: "" });
        }); }}>
          <h2>Invite a team member</h2>
          <div className="fields">
            <label className="field">Email<input type="email" required value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} /></label>
            <label className="field">Role<select value={f.role} onChange={(e) => setF({ ...f, role: e.target.value })}>{roles.map((r) => <option key={r}>{r}</option>)}</select></label>
          </div>
          <p className="muted small">Dispatchers run bookings and the board · finance sees reports and audit · viewers read only. Drivers are added from the Drivers tab.</p>
          <div><button className="btn primary" disabled={act.busy}>Create invite</button></div>
          {link && <div className="alert ok stack"><span>Send this one-time link (expires in 7 days):</span><CopyLink value={link} /></div>}
        </form>
      )}
      <ErrorBox error={act.error} />
      <div className="card">
        <h2>Team</h2>
        {!members ? <Spinner /> : (
          <table className="list">
            <thead><tr><th>Name</th><th>Role</th><th>Since</th></tr></thead>
            <tbody>
              {members.map((m) => (
                <tr key={m.userId}>
                  <td data-label="Name"><strong>{m.fullName}</strong><div className="muted small">{m.email}</div></td>
                  <td data-label="Role">{m.role}</td>
                  <td data-label="Since">{when(m.since)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
