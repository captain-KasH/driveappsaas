"use client";

import "leaflet/dist/leaflet.css";
import { useEffect, useRef } from "react";
import type * as L from "leaflet";

export type Pt = { lat: number; lng: number };
export type Marker = Pt & { label: string; color: string };

const COLOMBO: Pt = { lat: 6.9271, lng: 79.8612 };

// Map colours live here so the theme has one owner. Pickup blue / drop-off red are the
// convention passengers already know; teal is the Taxios brand (drivers, zones).
export const MAP_COLORS = { pickup: "#2563eb", dropoff: "#c62828", driver: "#0f766e", zone: "#0f766e", busy: "#c58a00", stale: "#9aa0a6" };

/**
 * Leaflet map with OpenStreetMap tiles. ponytail: OSM's public tiles and Nominatim are for low
 * volume only (their usage policies); swap for the chosen maps provider in T-013b / K-010.
 */
export default function MapView({
  markers = [], onPick, polygons = [], className = "map", fitKey,
}: {
  markers?: Marker[];
  onPick?: (p: Pt) => void;
  polygons?: GeoJSON.Geometry[];
  className?: string;
  fitKey?: string; // change to re-fit the view to markers/polygons
}) {
  const el = useRef<HTMLDivElement>(null);
  const map = useRef<L.Map | null>(null);
  const layer = useRef<L.LayerGroup | null>(null);
  const lib = useRef<typeof L | null>(null);
  const pick = useRef(onPick);
  pick.current = onPick;

  useEffect(() => {
    let cancelled = false;
    import("leaflet").then((Lmod) => {
      if (cancelled || !el.current || map.current) return;
      const Lx = (Lmod as any).default ?? Lmod;
      lib.current = Lx;
      const m = Lx.map(el.current, { zoomControl: true }).setView([COLOMBO.lat, COLOMBO.lng], 12);
      Lx.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
        maxZoom: 19,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      }).addTo(m);
      m.on("click", (e: L.LeafletMouseEvent) => pick.current?.({ lat: e.latlng.lat, lng: e.latlng.lng }));
      layer.current = Lx.layerGroup().addTo(m);
      map.current = m;
      draw(true);
    });
    return () => {
      cancelled = true;
      map.current?.remove();
      map.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const draw = (fit: boolean) => {
    const Lx = lib.current, g = layer.current, m = map.current;
    if (!Lx || !g || !m) return;
    g.clearLayers();
    const bounds: L.LatLngExpression[] = [];
    for (const poly of polygons) {
      const gj = Lx.geoJSON(poly as any, { style: { color: MAP_COLORS.zone, weight: 2, fillOpacity: 0.1 } }).addTo(g);
      const b = gj.getBounds();
      if (b.isValid()) bounds.push(b.getNorthEast(), b.getSouthWest());
    }
    for (const mk of markers) {
      Lx.circleMarker([mk.lat, mk.lng], { radius: 9, color: "#fff", weight: 2, fillColor: mk.color, fillOpacity: 1 })
        .bindTooltip(mk.label, { permanent: false })
        .addTo(g);
      bounds.push([mk.lat, mk.lng]);
    }
    if (fit && bounds.length) m.fitBounds(Lx.latLngBounds(bounds as any), { padding: [40, 40], maxZoom: 15 });
  };

  useEffect(() => draw(false), [markers, polygons]); // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => draw(true), [fitKey]); // eslint-disable-line react-hooks/exhaustive-deps

  return <div ref={el} className={className} role="application" aria-label="Map" />;
}

export type Place = Pt & { address: string };

const NOMINATIM = "https://nominatim.openstreetmap.org";

export async function searchPlaces(q: string): Promise<Place[]> {
  const url = `${NOMINATIM}/search?format=jsonv2&countrycodes=lk&limit=5&q=${encodeURIComponent(q)}`;
  const res = await fetch(url, { headers: { "Accept-Language": "en" } });
  if (!res.ok) throw new Error("Place search is unavailable right now. Tap the map instead.");
  const rows: any[] = await res.json();
  return rows.map((r) => ({ lat: Number(r.lat), lng: Number(r.lon), address: r.display_name }));
}

export async function reverseGeocode(p: Pt): Promise<string> {
  try {
    const res = await fetch(`${NOMINATIM}/reverse?format=jsonv2&zoom=17&lat=${p.lat}&lon=${p.lng}`, { headers: { "Accept-Language": "en" } });
    const r = await res.json();
    if (r?.display_name) return String(r.display_name).split(",").slice(0, 3).join(",");
  } catch {}
  return `${p.lat.toFixed(5)}, ${p.lng.toFixed(5)}`;
}

/** Circle polygon (GeoJSON, lng/lat) — used to draw a simple service zone from a centre and radius. */
export function circlePolygon(c: Pt, radiusKm: number, steps = 48): GeoJSON.Polygon {
  const coords: [number, number][] = [];
  const dLat = radiusKm / 110.574;
  const dLng = radiusKm / (111.32 * Math.cos((c.lat * Math.PI) / 180));
  for (let i = 0; i <= steps; i++) {
    const a = (i / steps) * 2 * Math.PI;
    coords.push([+(c.lng + dLng * Math.cos(a)).toFixed(6), +(c.lat + dLat * Math.sin(a)).toFixed(6)]);
  }
  coords[coords.length - 1] = coords[0];
  return { type: "Polygon", coordinates: [coords] };
}
