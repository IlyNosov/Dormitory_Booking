import type { Room } from "../types/bookings";

export const ROOM_ORDER: Room[] = [21, 256, 132];

export const ROOM_COLORS: Record<Room, { hex: string; ring: string; bg: string }> = {
    21: { hex: "#a78bfa", ring: "ring-violet-400/60", bg: "bg-violet-500/18" },
    256: { hex: "#f59e0b", ring: "ring-amber-400/60", bg: "bg-amber-500/18" },
    132: { hex: "#34d399", ring: "ring-emerald-400/60", bg: "bg-emerald-500/18" },
};
