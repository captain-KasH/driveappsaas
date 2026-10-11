// Taxios API client for the apps. Same contract as apps/web/lib/api.ts; the session token lives in
// the OS keychain (expo-secure-store), never in plain storage.
import Constants from "expo-constants";
import * as SecureStore from "expo-secure-store";

export const config = (Constants.expoConfig?.extra ?? {}) as { variant: "passenger" | "driver"; tenant: string; apiUrl: string };

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

const TOKEN_KEY = "taxios_session";
let cached: string | null | undefined;
let signedOut: (() => void) | null = null;

export const session = {
  async get() {
    if (cached === undefined) cached = await SecureStore.getItemAsync(TOKEN_KEY);
    return cached;
  },
  async set(t: string) {
    cached = t;
    await SecureStore.setItemAsync(TOKEN_KEY, t);
  },
  async clear() {
    cached = null;
    await SecureStore.deleteItemAsync(TOKEN_KEY);
  },
};

/** Called when the server rejects the session (expired, revoked from another device). */
export const onSignedOut = (fn: () => void) => { signedOut = fn; };

export const store = {
  get: (k: string) => SecureStore.getItemAsync(k),
  set: (k: string, v: string) => SecureStore.setItemAsync(k, v),
  del: (k: string) => SecureStore.deleteItemAsync(k),
};

async function request(method: string, path: string, init: RequestInit): Promise<Response> {
  const t = await session.get();
  try {
    return await fetch(config.apiUrl + path, { ...init, method, headers: { ...(init.headers as Record<string, string>), ...(t && { Authorization: `Bearer ${t}` }) } });
  } catch {
    throw new ApiError(0, "network", "Can't reach the server. Check your connection and try again.");
  }
}

export async function api<T = any>(method: string, path: string, body?: unknown, headers: Record<string, string> = {}): Promise<T> {
  const res = await request(method, path, {
    headers: { ...(body !== undefined && { "Content-Type": "application/json" }), ...headers },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && !path.startsWith("/v1/auth/")) {
      await session.clear();
      signedOut?.();
    }
    if (res.status === 429) throw new ApiError(429, "rate_limited", "Too many attempts. Wait a minute and try again.");
    throw new ApiError(res.status, data?.error?.code ?? "error", data?.error?.message ?? `Request failed (${res.status})`);
  }
  return data as T;
}

/** Uploads a photo/PDF picked on the device to POST {base}/files and returns its file id. */
export async function upload(base: string, file: { uri: string; name: string; type: string }): Promise<{ id: string; fileName: string }> {
  const body = new FormData();
  body.append("file", file as unknown as Blob);
  const res = await request("POST", `${base}/files`, { body });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new ApiError(res.status, data?.error?.code ?? "error", data?.error?.message ?? `Upload failed (${res.status})`);
  return data.file;
}

/** Idempotency key for bookings: retries of the same tap never double-book. */
export function newKey() {
  return "m-" + Date.now().toString(36) + "-" + Array.from({ length: 4 }, () => Math.random().toString(36).slice(2, 8)).join("");
}

export function money(minor: number | null | undefined, currency = "LKR") {
  if (minor == null) return "—";
  const v = (minor / 100).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  return `${currency} ${v}`;
}
export const km = (m: number | null | undefined) => (m == null ? "—" : `${(m / 1000).toFixed(1)} km`);
export const mins = (s: number | null | undefined) => (s == null ? "—" : `${Math.round(s / 60)} min`);
export function when(iso: string | null | undefined) {
  if (!iso) return "—";
  const d = new Date(iso);
  return `${d.toDateString().slice(0, 10)} ${d.toTimeString().slice(0, 5)}`;
}
export function ago(iso: string | null | undefined) {
  if (!iso) return "never";
  const s = Math.round((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.round(s / 60)} min ago`;
  return `${Math.round(s / 3600)} h ago`;
}

export type Membership = { tenantId: string; tenantSlug: string; tenantName: string; tenantStatus: string; role: string };
export type Me = { user: { id: string; email: string; fullName: string }; memberships: Membership[] };
export type Brand = { name: string; color: string; onColor: string; logoUrl: string | null; tagline: string | null; supportPhone: string | null; supportEmail: string | null; websiteUrl: string | null; termsUrl: string | null; privacyUrl: string | null; currency: string };
export type PublicTenant = { tenant: { id: string; name: string }; brand: Brand; categories: { id: string; name: string; seats: number }[] };
export type Breakdown = { currency: string; lines: { code: string; label: string; amountMinor: number }[]; taxMinor: number; totalMinor: number };

export async function signIn(email: string, password: string) {
  const r = await api<{ token: string }>("POST", "/v1/auth/login", { email: email.trim(), password, client: "mobile" });
  await session.set(r.token);
}
export async function signOut() {
  await api("POST", "/v1/auth/logout").catch(() => {});
  await session.clear();
}
