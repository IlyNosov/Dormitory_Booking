import type { RoomInfo } from "../types/bookings";

const BASE = "/api";

async function readError(r: Response): Promise<string> {
  try { return (await r.text()).trim() || `${r.status}`; }
  catch { return String(r.status); }
}

export async function fetchRooms(): Promise<RoomInfo[]> {
  const r = await fetch(`${BASE}/rooms`, { credentials: "include" });
  if (!r.ok) return [];
  const data = await r.json();
  if (!Array.isArray(data)) return [];
  // API returns PascalCase — normalize to camelCase
  return data.map((d: any) => ({
    number:       d.Number       ?? d.number,
    name:         d.Name         ?? d.name,
    isActive:     d.IsActive     ?? d.isActive ?? true,
    weekdayOpen:  d.WeekdayOpen  ?? d.weekdayOpen  ?? 6,
    weekdayClose: d.WeekdayClose ?? d.weekdayClose ?? 23,
    friSatOpen:   d.FriSatOpen   ?? d.friSatOpen   ?? 6,
    friSatClose:  d.FriSatClose  ?? d.friSatClose  ?? 25,
    sunOpen:      d.SunOpen      ?? d.sunOpen      ?? 6,
    sunClose:     d.SunClose     ?? d.sunClose      ?? 23,
  }));
}

export async function createRoom(number: number, name: string): Promise<RoomInfo> {
  const r = await fetch(`${BASE}/rooms`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ number, name }),
  });
  if (!r.ok) throw new Error(await readError(r));
  return r.json();
}

export async function deleteRoom(number: number): Promise<void> {
  const r = await fetch(`${BASE}/rooms/${number}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!r.ok && r.status !== 204) throw new Error(await readError(r));
}
