import type { AdminEntry } from "../types/bookings";

const BASE = "/api";

async function readError(r: Response): Promise<string> {
  try { return (await r.text()).trim() || `${r.status}`; }
  catch { return String(r.status); }
}

export async function fetchAdmins(): Promise<AdminEntry[]> {
  const r = await fetch(`${BASE}/admins`, { credentials: "include" });
  if (!r.ok) return [];
  const data = await r.json();
  return Array.isArray(data) ? data : [];
}

export async function addAdmin(email: string): Promise<void> {
  const r = await fetch(`${BASE}/admins`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!r.ok) throw new Error(await readError(r));
}

export async function removeAdmin(email: string): Promise<void> {
  const r = await fetch(`${BASE}/admins`, {
    method: "DELETE",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!r.ok && r.status !== 204) throw new Error(await readError(r));
}
