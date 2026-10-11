import type { Metadata, Viewport } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Taxios — ride-hailing software for Sri Lankan transport businesses",
  description: "Run your own branded taxi, tuk and transfer service: bookings, dispatch, drivers and fares in one platform.",
};

export const viewport: Viewport = { width: "device-width", initialScale: 1, themeColor: "#0f766e" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
