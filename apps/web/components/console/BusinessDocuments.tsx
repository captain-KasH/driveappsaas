"use client";

import { useCallback, useEffect, useState } from "react";
import { api, openFile, when } from "@/lib/api";
import { ErrorBox, useAction } from "@/components/ui";
import FilePick from "@/components/FilePick";
import { Ctx, can } from "./ctx";

type Doc = { id: string; docType: string; fileId: string; fileName: string; note: string | null; createdAt: string };
const TYPES: [string, string][] = [["business_registration", "Business registration certificate"], ["owner_id", "Owner's NIC / passport"], ["tax_registration", "Tax registration"], ["other", "Other"]];

/** Verification documents Taxios reviews before a self-registered business goes live (T-122). */
export default function BusinessDocuments({ ctx, onChange }: { ctx: Ctx; onChange: () => Promise<void> }) {
  const base = `/v1/tenants/${ctx.tenantId}`;
  const editable = can(ctx.role, "tenant.manage");
  const [docs, setDocs] = useState<Doc[] | null>(null);
  const [docType, setDocType] = useState("business_registration");
  const [file, setFile] = useState<{ id: string; fileName: string } | null>(null);
  const [note, setNote] = useState("");
  const act = useAction();
  const load = useCallback(() => api<{ documents: Doc[] }>("GET", `${base}/documents`).then((r) => setDocs(r.documents)).catch(act.setError), [base, act.setError]);
  useEffect(() => { load(); }, [load]);

  return (
    <div className="card stack">
      <h2>Verification documents</h2>
      <p className="muted small">{ctx.tenant.verifiedAt ? "This business is verified." : "Taxios checks the registration certificate before your booking page goes public. Only Taxios staff and your admins can open these files."}</p>
      {docs && docs.length > 0 && (
        <table className="list"><tbody>
          {docs.map((d) => (
            <tr key={d.id}>
              <td data-label="Document"><strong>{TYPES.find(([k]) => k === d.docType)?.[1] ?? d.docType}</strong>{d.note && <div className="muted small">{d.note}</div>}</td>
              <td data-label="File"><button className="linkbtn" onClick={() => openFile(`${base}/files/${d.fileId}`).catch(act.setError)}>{d.fileName}</button></td>
              <td data-label="Added">{when(d.createdAt)}</td>
            </tr>
          ))}
        </tbody></table>
      )}
      {editable && (
        <form className="stack" onSubmit={(e) => {
          e.preventDefault();
          act.run(async () => {
            if (!file) throw new Error("Choose a file first.");
            await api("POST", `${base}/documents`, { docType, fileId: file.id, note });
            setFile(null); setNote("");
            await load(); await onChange();
          });
        }}>
          <div className="fields">
            <label className="field">Type<select value={docType} onChange={(e) => setDocType(e.target.value)}>{TYPES.map(([v, l]) => <option key={v} value={v}>{l}</option>)}</select></label>
            <FilePick key={file?.id ?? "empty"} base={base} value={file} onChange={setFile} />
            <label className="field">Note (optional)<input value={note} maxLength={500} onChange={(e) => setNote(e.target.value)} placeholder="e.g. PV 12345, issued 2024" /></label>
          </div>
          <ErrorBox error={act.error} />
          <div><button className="btn primary" disabled={act.busy || !file}>Add document</button></div>
        </form>
      )}
    </div>
  );
}
