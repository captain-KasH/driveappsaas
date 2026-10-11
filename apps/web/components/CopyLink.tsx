"use client";

import { useState } from "react";

export function CopyLink({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="row" style={{ flexWrap: "nowrap", gap: 6 }}>
      <input readOnly value={value} onFocus={(e) => e.currentTarget.select()} className="mono small" aria-label="Link" />
      <button type="button" className="btn small" onClick={async () => {
        await navigator.clipboard.writeText(value).catch(() => {});
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      }}>{copied ? "Copied" : "Copy"}</button>
    </div>
  );
}
