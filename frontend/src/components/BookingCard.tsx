import React, { useEffect, useState } from "react";
import { Lock, Trash2, AlignLeft, Send } from "lucide-react";
import type { Booking } from "../types/bookings";
import { getRoomColor } from "../utils/rooms";
import { RoomIcon } from "./RoomIcon";
import { CardRoomArt } from "./CardRoomArt";
import { msToHuman } from "../utils/date";

function toHM(d: string) {
  return new Date(d).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

export function displayRoom(r: number): string {
  if (r === 2812) return "812 (2к)";
  if (r === 3812) return "812 (3к)";
  return String(r);
}

interface Props {
  booking: Booking;
  onDelete?: () => void;
  onInspect?: () => void;
  showDate?: boolean;
}

export function BookingCard({ booking: b, onDelete, onInspect, showDate }: Props) {
  const rc = getRoomColor(b.room);

  const [, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 60000);
    return () => clearInterval(id);
  }, []);

  const now = new Date();
  const start = new Date(b.start);
  const end = new Date(b.end);
  const isPast = end < now;
  const isActive = start <= now && now <= end;
  const progress = isActive
    ? (now.getTime() - start.getTime()) / (end.getTime() - start.getTime())
    : 0;
  const remainingMs = isActive ? end.getTime() - now.getTime() : 0;
  const remainingText = msToHuman(remainingMs);

  const dateLabel = showDate
    ? new Date(b.start).toLocaleDateString("ru-RU", { day: "2-digit", month: "long" })
    : null;

  return (
    // flex flex-col so content div can push bottom section to the card bottom
    <div
      className={[
        "group relative rounded-2xl overflow-hidden cursor-pointer border transition-all duration-200 ease-out hover:shadow-md hover:-translate-y-0.5 flex flex-col",
        isPast ? "opacity-50 grayscale-[0.3]" : "",
      ].join(" ")}
      style={{
        background: "var(--d-surface)",
        borderColor: rc.hex + "55",
        boxShadow: "0 1px 4px rgba(0,0,0,0.05)",
      }}
      onClick={onInspect}
    >
      {/* Room art illustration on right */}
      <CardRoomArt room={b.room} />

      {/* Top accent bar */}
      <div className="h-1 w-full relative z-[1] shrink-0" style={{ background: rc.hex }} />

      {/* Content — flex-col, flex-1 so it fills card height and pushes bottom down */}
      <div className="p-4 flex flex-col gap-2 relative z-[1] flex-1">

        {/* Title + status icons */}
        <div className="flex items-start justify-between gap-2">
          <div
            className="font-semibold leading-tight line-clamp-2 text-sm min-w-0"
            style={{ color: "var(--d-text)" }}
          >
            {b.title}
          </div>
          <div className="flex items-center gap-1.5 shrink-0 mt-0.5">
            {isActive && (
              <span className="inline-block w-2 h-2 rounded-full bg-[var(--d-success)] animate-pulse" />
            )}
            {b.isPrivate && (
              <Lock className="h-3.5 w-3.5" style={{ color: "var(--d-secondary)" }} />
            )}
            {b.description && (
              <AlignLeft className="h-3.5 w-3.5" style={{ color: "var(--d-text-muted)" }} />
            )}
          </div>
        </div>

        {/* Time */}
        <div className="text-sm font-medium" style={{ color: "var(--d-text-sec)" }}>
          {dateLabel && <span className="mr-1">{dateLabel} •</span>}
          {toHM(b.start)} — {toHM(b.end)}
        </div>

        {/* Active indicator — static text, no animation (prevents jerk on re-render) */}
        {isActive && (
          <div className="flex items-center gap-1.5 text-xs" style={{ color: "var(--d-success)" }}>
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-[var(--d-success)] animate-pulse shrink-0" />
            <span>Идёт · {remainingText}</span>
          </div>
        )}

        {/* Telegram — truncated */}
        {b.telegramId && (
          <div className="flex items-center gap-1 text-xs min-w-0 overflow-hidden" style={{ color: "var(--d-text-muted)" }}>
            <Send className="h-3 w-3 shrink-0" />
            <span className="truncate">@{b.telegramId}</span>
          </div>
        )}

        {/* Bottom row — mt-auto pushes it to card bottom even when card is expanded */}
        <div className="mt-auto flex items-center justify-between gap-2">
          <span
            className="flex items-center gap-1 text-xs px-2.5 py-1 rounded-full font-medium shrink-0"
            style={{ background: rc.hex + "22", color: rc.hex }}
          >
            <RoomIcon roomNumber={b.room} size={14} />
            Комната {displayRoom(b.room)}
          </span>

          {onDelete && (
            <button
              className="opacity-0 group-hover:opacity-100 inline-flex items-center justify-center rounded-lg p-1.5 transition-all duration-200 shrink-0"
              style={{ color: "var(--d-error)" }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "rgba(192,117,106,0.12)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
              onClick={(e) => { e.stopPropagation(); onDelete(); }}
              title="Удалить"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Active progress bar */}
        {isActive && (
          <div className="h-1 rounded-full overflow-hidden" style={{ background: "var(--d-border)" }}>
            <div
              className="h-full rounded-full transition-all duration-1000"
              style={{ width: `${progress * 100}%`, background: "var(--d-success)" }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
