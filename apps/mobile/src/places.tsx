import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";
import * as Location from "expo-location";
import { Field, P, c } from "./ui";

export type Place = { lat: number; lng: number; address: string };

// ponytail: OpenStreetMap Nominatim, same as the web (K-014) — low-volume use only; swap for the
// chosen maps provider in T-013b. Their policy requires an identifying User-Agent.
const NOMINATIM = "https://nominatim.openstreetmap.org";
const HEADERS = { "Accept-Language": "en", "User-Agent": "Taxios/0.1 (taxios.packline.lk)" };

export async function searchPlaces(q: string): Promise<Place[]> {
  const res = await fetch(`${NOMINATIM}/search?format=jsonv2&countrycodes=lk&limit=5&q=${encodeURIComponent(q)}`, { headers: HEADERS });
  if (!res.ok) throw new Error("Place search is unavailable right now.");
  const rows: any[] = await res.json();
  return rows.map((r) => ({ lat: Number(r.lat), lng: Number(r.lon), address: String(r.display_name).split(",").slice(0, 3).join(",") }));
}

export async function reverseGeocode(lat: number, lng: number): Promise<string> {
  try {
    const res = await fetch(`${NOMINATIM}/reverse?format=jsonv2&zoom=17&lat=${lat}&lon=${lng}`, { headers: HEADERS });
    const r = await res.json();
    if (r?.display_name) return String(r.display_name).split(",").slice(0, 3).join(",");
  } catch {}
  return `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
}

/** One-shot current position. Throws a readable message on denial — typing an address always works. */
export async function currentPlace(): Promise<Place> {
  const perm = await Location.requestForegroundPermissionsAsync();
  if (perm.status !== "granted") throw new Error("Location is off for this app. Type your pickup address instead.");
  const p = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.Balanced });
  return { lat: p.coords.latitude, lng: p.coords.longitude, address: await reverseGeocode(p.coords.latitude, p.coords.longitude) };
}

/** Address search box with debounced suggestions. */
export function PlaceSearch({ label, value, onChange, placeholder }: { label: string; value: Place | null; onChange: (p: Place | null) => void; placeholder?: string }) {
  const [q, setQ] = useState(value?.address ?? "");
  const [results, setResults] = useState<Place[]>([]);
  const [err, setErr] = useState<string | null>(null);
  useEffect(() => { setQ(value?.address ?? ""); }, [value]);
  useEffect(() => {
    if (value && q === value.address) return;
    if (q.trim().length < 3) { setResults([]); return; }
    const id = setTimeout(() => searchPlaces(q).then((r) => { setResults(r); setErr(r.length ? null : "No matches. Try a nearby landmark."); }).catch((e) => setErr(e.message)), 450);
    return () => clearTimeout(id);
  }, [q, value]);
  return (
    <View style={{ gap: 6 }}>
      <Field label={label} value={q} placeholder={placeholder} onChangeText={(t) => { setQ(t); if (value) onChange(null); }} autoCorrect={false} />
      {err && !value ? <P muted small>{err}</P> : null}
      {!value && results.map((r) => (
        <Pressable key={`${r.lat},${r.lng}`} accessibilityRole="button" onPress={() => { onChange(r); setResults([]); }}
          style={{ padding: 10, backgroundColor: c.surface, borderRadius: 8, borderWidth: 1, borderColor: c.line }}>
          <Text style={{ color: c.text }}>{r.address}</Text>
        </Pressable>
      ))}
    </View>
  );
}
