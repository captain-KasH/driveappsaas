"use client";

import { useState } from "react";
import { Place, searchPlaces } from "./MapView";

/** Address search (Sri Lanka). Typing an address never requires location permission. */
export default function PlaceInput({ label, value, onChange, active, onFocus }: {
  label: string;
  value: Place | null;
  onChange: (p: Place) => void;
  active?: boolean;
  onFocus?: () => void;
}) {
  const [q, setQ] = useState("");
  const [results, setResults] = useState<Place[]>([]);
  const [msg, setMsg] = useState("");

  const search = async () => {
    if (q.trim().length < 3) return;
    setMsg("Searching…");
    try {
      const r = await searchPlaces(q);
      setResults(r);
      setMsg(r.length ? "" : "No matches. Try a nearby landmark, or tap the map.");
    } catch (e) {
      setMsg((e as Error).message);
    }
  };

  return (
    <div className="stack" style={{ gap: 6 }}>
      <label className="field">
        {label}
        <div className="row" style={{ flexWrap: "nowrap", gap: 6 }}>
          <input
            value={q}
            placeholder={value?.address ?? "Search a place or tap the map"}
            onFocus={onFocus}
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), search())}
            style={active ? { borderColor: "var(--accent)" } : undefined}
          />
          <button type="button" className="btn small" onClick={search}>Find</button>
        </div>
      </label>
      {value && <div className="small">📍 {value.address}</div>}
      {msg && <div className="small muted">{msg}</div>}
      {results.length > 0 && (
        <div className="card" style={{ padding: 4 }}>
          {results.map((r, i) => (
            <button key={i} type="button" className="btn small block" style={{ justifyContent: "flex-start", border: 0, textAlign: "left" }}
              onClick={() => { onChange({ ...r, address: r.address.split(",").slice(0, 3).join(",") }); setResults([]); setQ(""); }}>
              {r.address}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
