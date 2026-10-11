import Link from "next/link";
import { TopBar } from "@/components/ui";

export default function Landing() {
  return (
    <>
      <TopBar>
        <Link className="btn small" href="/login/">Sign in</Link>
      </TopBar>
      <main className="page">
        <section className="hero">
          <h1>Run your own ride-hailing service, under your own name.</h1>
          <p>
            Taxios gives Sri Lankan taxi, tuk, airport-transfer and chauffeur businesses the whole stack:
            passenger booking, dispatch, driver checks, server-calculated fares and trip records.
          </p>
          <div className="row">
            <Link className="btn primary" href="/login/">Sign in to your business</Link>
            <a className="btn" href="mailto:hello@packline.lk?subject=Taxios%20for%20my%20business">Talk to us</a>
          </div>
        </section>
        <div className="grid">
          {[
            ["Dispatch that holds up", "Nearest eligible driver gets an exclusive offer. Nobody double-books a ride, even when two taps land at once."],
            ["Fares you set, computed by the server", "Base, distance, time, minimum, waiting, night and scheduled fees. Every booking keeps the fare version it was quoted."],
            ["Drivers you can vouch for", "Licence, ID and insurance are reviewed before a driver can go online, and expired documents stop dispatch automatically."],
            ["Your brand, your passengers", "Passengers book on your page. Your data is isolated from every other business on the platform."],
          ].map(([t, d]) => (
            <div className="card" key={t}>
              <h3>{t}</h3>
              <p className="muted small">{d}</p>
            </div>
          ))}
        </div>
        <p className="muted small" style={{ marginTop: 32 }}>
          Booking with a Taxios operator? Use the link your operator shared — it looks like <code>taxios.packline.lk/book/?b=your-operator</code>.
        </p>
      </main>
    </>
  );
}
