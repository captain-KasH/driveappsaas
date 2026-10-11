"use client";

import { useRef, useState } from "react";
import { upload } from "@/lib/api";

/** Picks a JPEG/PNG/WebP/PDF ≤ 10 MB, uploads it immediately and hands back the file id. */
export default function FilePick({ base, value, onChange, label = "Photo or PDF" }: {
  base: string; value: { id: string; fileName: string } | null; onChange: (f: { id: string; fileName: string } | null) => void; label?: string;
}) {
  const input = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  return (
    <label className="field">{label}
      <div className="row" style={{ gap: 8 }}>
        <input ref={input} type="file" accept="image/jpeg,image/png,image/webp,application/pdf" disabled={busy}
          onChange={async (e) => {
            const f = e.target.files?.[0];
            if (!f) return;
            if (f.size > 10 * 1024 * 1024) { setErr("Files must be 10 MB or smaller."); return; }
            setBusy(true); setErr(null);
            try { const r = await upload(`${base}/files`, f); onChange({ id: r.id, fileName: r.fileName }); }
            catch (x) { setErr(x instanceof Error ? x.message : String(x)); onChange(null); }
            finally { setBusy(false); }
          }} />
        {busy && <span className="muted small">Uploading…</span>}
        {value && !busy && <span className="badge ok">{value.fileName}</span>}
      </div>
      {err && <span className="small" style={{ color: "var(--bad)" }}>{err}</span>}
    </label>
  );
}
