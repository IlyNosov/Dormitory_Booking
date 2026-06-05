import type { Room } from "../types/bookings";

// Nature/forest palette for rooms
export const ROOM_COLORS: Record<number, { hex: string; ring: string; bg: string }> = {
  21:   { hex: "#578f64", ring: "ring-[#578f64]/60",  bg: "bg-[#578f64]/[0.12]" },  // green (TV room)
  132:  { hex: "#869c27", ring: "ring-[#869c27]/60",  bg: "bg-[#869c27]/[0.12]" },  // olive (couch room)
  256:  { hex: "#795638", ring: "ring-[#795638]/60",  bg: "bg-[#795638]/[0.12]" },  // brown (piano room)
  2812: { hex: "#c0915f", ring: "ring-[#c0915f]/60",  bg: "bg-[#c0915f]/[0.12]" },  // amber (coworking 2к)
  3812: { hex: "#c1ac6c", ring: "ring-[#c1ac6c]/60",  bg: "bg-[#c1ac6c]/[0.12]" },  // gold (coworking 3к)
};

const PALETTE = [
  { hex: "#5B7A3F", ring: "ring-[#5B7A3F]/60", bg: "bg-[#5B7A3F]/[0.12]" },
  { hex: "#C49A5E", ring: "ring-[#C49A5E]/60", bg: "bg-[#C49A5E]/[0.12]" },
  { hex: "#7A8569", ring: "ring-[#7A8569]/60", bg: "bg-[#7A8569]/[0.12]" },
  { hex: "#3A5428", ring: "ring-[#3A5428]/60", bg: "bg-[#3A5428]/[0.12]" },
  { hex: "#9B6B3E", ring: "ring-[#9B6B3E]/60", bg: "bg-[#9B6B3E]/[0.12]" },
];

export function getRoomColor(room: number) {
  if (ROOM_COLORS[room]) return ROOM_COLORS[room];
  return PALETTE[room % PALETTE.length];
}
