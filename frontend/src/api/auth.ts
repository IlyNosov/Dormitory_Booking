import type { User } from "../types/bookings";

const BASE = "/api";

async function readError(r: Response): Promise<string> {
  try { return (await r.text()).trim() || `${r.status}`; }
  catch { return String(r.status); }
}

export async function requestOTP(email: string): Promise<void> {
  const r = await fetch(`${BASE}/auth/otp/request`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!r.ok) throw new Error(await readError(r));
}

export async function verifyOTP(email: string, code: string): Promise<User> {
  const r = await fetch(`${BASE}/auth/otp/verify`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, code }),
  });
  if (!r.ok) throw new Error(await readError(r));
  return r.json();
}

export async function getMe(): Promise<User | null> {
  const r = await fetch(`${BASE}/auth/me`, { credentials: "include" });
  if (r.status === 401) return null;
  if (!r.ok) return null;
  return r.json();
}

export async function logout(): Promise<void> {
  await fetch(`${BASE}/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}
