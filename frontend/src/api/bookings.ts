import type { Booking, CreateBookingPayload, UpdateBookingPayload } from "../types/bookings";

const BASE = "/api";

async function readError(r: Response): Promise<string> {
  try { return (await r.text()).trim() || `${r.status}`; }
  catch { return String(r.status); }
}

function mapBooking(raw: Record<string, unknown>): Booking {
  return {
    id: raw.id as string,
    start: (raw.start ?? raw.startAt) as string,
    end: (raw.end ?? raw.endAt) as string,
    room: raw.room as number,
    title: raw.title as string,
    isPrivate: raw.isPrivate as boolean,
    description: raw.description as string | undefined,
    userEmail: raw.userEmail as string | undefined,
    canManage: raw.canManage as boolean | undefined,
    telegramId: (raw.telegramId ?? raw.telegram_id) as string | undefined,
  };
}

export async function fetchBookings(): Promise<Booking[]> {
  const r = await fetch(`${BASE}/bookings`, { credentials: "include" });
  if (!r.ok) return [];
  const data = await r.json();
  return Array.isArray(data) ? data.map(mapBooking) : [];
}

export async function createBooking(payload: CreateBookingPayload): Promise<Booking> {
  const r = await fetch(`${BASE}/bookings`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(await readError(r));
  return r.json();
}

export async function forceCreateBooking(payload: CreateBookingPayload): Promise<Booking> {
  const r = await fetch(`${BASE}/bookings/force`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(await readError(r));
  return r.json();
}

export async function updateBooking(id: string, payload: UpdateBookingPayload): Promise<Booking> {
  const r = await fetch(`${BASE}/bookings/${id}`, {
    method: "PATCH",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error(await readError(r));
  return mapBooking(await r.json());
}

export async function deleteBooking(id: string): Promise<void> {
  const r = await fetch(`${BASE}/bookings/${id}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!r.ok && r.status !== 204) throw new Error(await readError(r));
}
