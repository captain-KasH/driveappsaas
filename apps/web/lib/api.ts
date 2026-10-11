// Thin client for the Taxios API (same origin). Errors carry the server's code and message.

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

const TOKEN_KEY = "taxios_token";

// ponytail: bearer token in localStorage; move to an httpOnly cookie before third-party scripts are added.
export const token = {
  get: () => (typeof window === "undefined" ? null : window.localStorage.getItem(TOKEN_KEY)),
  set: (t: string) => window.localStorage.setItem(TOKEN_KEY, t),
  clear: () => window.localStorage.removeItem(TOKEN_KEY),
};

export async function api<T = any>(method: string, path: string, body?: unknown, headers: Record<string, string> = {}): Promise<T> {
  const t = token.get();
  let res: Response;
  try {
    res = await fetch(path, {
      method,
      headers: {
        ...(body !== undefined && { "Content-Type": "application/json" }),
        ...(t && { Authorization: `Bearer ${t}` }),
        ...headers,
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiError(0, "network", "Can't reach Taxios. Check your connection and try again.");
  }
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && t && !path.startsWith("/v1/auth/")) token.clear();
    if (res.status === 429) throw new ApiError(429, "rate_limited", "Too many attempts. Wait a minute and try again.");
    throw new ApiError(res.status, data?.error?.code ?? "error", data?.error?.message ?? `Request failed (${res.status})`);
  }
  return data as T;
}

export const newIdempotencyKey = () => crypto.randomUUID();

export function money(minor: number | null | undefined, currency = "LKR") {
  if (minor == null) return "—";
  return `${currency} ${(minor / 100).toLocaleString("en-LK", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function when(iso: string | null | undefined) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-LK", { weekday: "short", day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });
}

export function ago(iso: string | null | undefined) {
  if (!iso) return "never";
  const s = Math.round((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.round(s / 60)} min ago`;
  return `${Math.round(s / 3600)} h ago`;
}

export const km = (m: number | null | undefined) => (m == null ? "—" : `${(m / 1000).toFixed(1)} km`);
export const mins = (s: number | null | undefined) => (s == null ? "—" : `${Math.round(s / 60)} min`);

export type Membership = { tenantId: string; tenantSlug: string; tenantName: string; tenantStatus: string; role: string };
export type Me = { user: { id: string; email: string; fullName: string; platformRole: string; emailVerified: boolean }; memberships: Membership[] };

/** Uploads one file (multipart) and returns the stored file record. */
export async function upload(path: string, file: File): Promise<{ id: string; contentType: string; sizeBytes: number; fileName: string }> {
  const t = token.get();
  const body = new FormData();
  body.append("file", file, file.name);
  let res: Response;
  try {
    res = await fetch(path, { method: "POST", headers: { ...(t && { Authorization: `Bearer ${t}` }) }, body });
  } catch {
    throw new ApiError(0, "network", "Can't reach Taxios. Check your connection and try again.");
  }
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new ApiError(res.status, data?.error?.code ?? "error", data?.error?.message ?? `Upload failed (${res.status})`);
  return data.file;
}

/** Opens an authenticated file in a new tab (the bearer token can't ride on a plain link). */
export async function openFile(path: string) {
  const t = token.get();
  const res = await fetch(path, { headers: { ...(t && { Authorization: `Bearer ${t}` }) } });
  if (!res.ok) throw new ApiError(res.status, "file_unavailable", "Couldn't open this file.");
  const url = URL.createObjectURL(await res.blob());
  window.open(url, "_blank", "noopener");
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}
