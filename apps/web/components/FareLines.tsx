import { money } from "@/lib/api";

export type Breakdown = { currency: string; lines: { code: string; label: string; amountMinor: number }[]; subtotalMinor: number; taxMinor: number; totalMinor: number };

export default function FareLines({ b, title }: { b: Breakdown | null | undefined; title?: string }) {
  if (!b) return null;
  return (
    <table className="fare-lines">
      {title && <caption className="muted small" style={{ textAlign: "left" }}>{title}</caption>}
      <tbody>
        {b.lines.map((l) => <tr key={l.code}><td>{l.label}</td><td>{money(l.amountMinor, b.currency)}</td></tr>)}
        {b.taxMinor > 0 && <tr><td>Tax</td><td>{money(b.taxMinor, b.currency)}</td></tr>}
        <tr className="total"><td>Total</td><td>{money(b.totalMinor, b.currency)}</td></tr>
      </tbody>
    </table>
  );
}
