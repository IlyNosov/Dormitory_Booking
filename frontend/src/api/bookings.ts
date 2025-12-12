import type { Bookings, CreateBookingPayload } from "../types/bookings";
import { normalizeApiError } from "../utils/errors";

const API_BASE = (import.meta as any).env?.VITE_API_BASE || "";

let ADMIN_TOKEN = "";

export function setAdminToken(t: string | null) {
    ADMIN_TOKEN = (t || "").trim();
}

function withHeaders(extra?: Record<string, string>) {
    return {
        "Content-Type": "application/json",
        "X-User-Email": "student@edu.hse.ru",
        ...(ADMIN_TOKEN ? { "X-Admin-Token": ADMIN_TOKEN } : {}),
        ...(extra ?? {}),
    };
}

async function readErrorText(r: Response): Promise<string> {
    const ct = r.headers.get("content-type") || "";
    try {
        if (ct.includes("application/json")) {
            const j = await r.json();
            const msg = j?.error || j?.message || JSON.stringify(j);
            return typeof msg === "string" ? msg : String(msg);
        }
    } catch {
        // ignore
    }
    try {
        return await r.text();
    } catch {
        return `${r.status} ${r.statusText}`;
    }
}

export async function fetchBookings(): Promise<Bookings[]> {
    const r = await fetch(`${API_BASE}/api/bookings`, {
        credentials: "include",
        headers: withHeaders(),
    });
    if (!r.ok) return [];
    const data = (await r.json()) as Bookings[];
    return Array.isArray(data) ? data : [];
}

export async function createBooking(payload: CreateBookingPayload): Promise<Bookings> {
    const r = await fetch(`${API_BASE}/api/bookings`, {
        method: "POST",
        credentials: "include",
        headers: withHeaders({}),
        body: JSON.stringify(payload),
    });
    if (!r.ok) throw new Error(await r.text());
    return (await r.json()) as Bookings;
}

export async function deleteBooking(id: string, telegramId?: string): Promise<void> {
    const url = new URL(`${API_BASE}/api/bookings/${id}`, window.location.origin);
    if (telegramId) url.searchParams.set("tg", telegramId);

    const r = await fetch(url.toString(), {
        method: "DELETE",
        credentials: "include",
        headers: withHeaders({}),
    });
    if (!r.ok && r.status !== 204) throw new Error(await r.text());
}