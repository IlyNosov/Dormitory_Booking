import React, { useEffect, useMemo, useState } from "react";
import { getRoomColor } from "../utils/rooms";
import type { Booking } from "../types/bookings";

const DEFAULT_ROOM_ORDER = [21, 256, 132, 2812, 3812] as const;
const PX_PER_HOUR = 56;
const TIME_COL_W = 72;
const HEADER_H = 37;

function dayLimits(_d: Date) {
  return { startH: 6, endH: 25 };
}

function displayHour(h: number): string {
  const real = h % 24;
  return `${String(real).padStart(2, "0")}:00`;
}

function toHM(dt: string) {
  const d = new Date(dt);
  return d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

function useNowTop(dayKey: string, startH: number): number | null {
  const [top, setTop] = useState<number | null>(null);

  useEffect(() => {
    function update() {
      const now = new Date();
      const d = new Date(dayKey + "T00:00:00");
      if (
        now.getFullYear() !== d.getFullYear() ||
        now.getMonth() !== d.getMonth() ||
        now.getDate() !== d.getDate()
      ) {
        setTop(null);
        return;
      }
      const h = now.getHours() + now.getMinutes() / 60 + now.getSeconds() / 3600;
      setTop((h - startH) * PX_PER_HOUR);
    }
    update();
    const id = setInterval(update, 30_000);
    return () => clearInterval(id);
  }, [dayKey, startH]);

  return top;
}

export function BookingsTableView({
  dayKey,
  bookings,
  onInspect,
}: {
  dayKey: string;
  bookings: Booking[];
  onInspect: (b: Booking) => void;
}) {
  const d = useMemo(() => new Date(dayKey + "T00:00:00"), [dayKey]);
  const { startH, endH } = useMemo(() => dayLimits(d), [d]);
  const totalH = (endH - startH) * PX_PER_HOUR;
  // endH is 1:00 AM next day; exclude it from labels to avoid "01:00" and bottom overflow
  const hours = Array.from({ length: endH - startH }, (_, i) => startH + i);

  const nowTop = useNowTop(dayKey, startH);

  const roomOrder = useMemo(() => {
    const present = new Set(bookings.map(b => b.room));
    const base = [...DEFAULT_ROOM_ORDER] as number[];
    const extras = [...present].filter(r => !base.includes(r)).sort((a, b) => a - b);
    return [...base, ...extras];
  }, [bookings]);

  const byRoom = useMemo(() => {
    const map = new Map<number, Booking[]>();
    for (const r of roomOrder) map.set(r, []);
    for (const b of bookings) {
      const list = map.get(b.room);
      if (list) list.push(b);
    }
    for (const [, v] of map) v.sort((a, b) => +new Date(a.start) - +new Date(b.start));
    return map;
  }, [bookings, roomOrder]);

  return (
    <section
      className="rounded-2xl border overflow-hidden"
      style={{
        background: "var(--d-surface)",
        borderColor: "var(--d-border)",
        boxShadow: "0 1px 4px rgba(0,0,0,0.05)",
      }}
    >
      {/*
        Single scroll container: owns both H and V scroll.
        - Fixes PC "doesn't scroll immediately": overflow-x:auto alone converts
          overflow-y:visible→auto and traps wheel events; single overflow:auto
          scrolls on hover without any click.
        - Fixes mobile over-pull: overscrollBehavior contain.
        - Sticky time column (left:0) and sticky header row (top:0) work
          correctly inside a single scroll ancestor.
      */}
      <div
        style={{
          overflow: "auto",
          maxHeight: "calc(100dvh - 220px)",
          minHeight: 300,
          // contain: bounce stays inside this element; native momentum in both axes.
          overscrollBehavior: "contain",
          scrollbarWidth: "thin",
          // Same colour as the table so the elastic overscroll area isn't white.
          background: "var(--d-surface)",
        } as React.CSSProperties}
      >
        {/* Wide content — at least TIME_COL_W + rooms × 160 px */}
        <div style={{ minWidth: TIME_COL_W + roomOrder.length * 160, position: "relative" }}>

          {/* ── Sticky header row (room names) ─────────────────────────────── */}
          <div
            className="flex border-b"
            style={{
              position: "sticky",
              top: 0,
              zIndex: 20,
              borderColor: "var(--d-border)",
              background: "var(--d-panel)",
            }}
          >
            {/* Corner cell — sticky left inside the sticky row */}
            <div
              style={{
                width: TIME_COL_W,
                minWidth: TIME_COL_W,
                flexShrink: 0,
                height: HEADER_H,
                borderRight: "1px solid var(--d-border)",
                background: "var(--d-panel)",
                position: "sticky",
                left: 0,
                zIndex: 1,
              }}
            />

            {roomOrder.map((room) => {
              const rc = getRoomColor(room);
              return (
                <div
                  key={room}
                  className="flex-1 flex items-center justify-center px-2 text-xs font-semibold uppercase tracking-wide border-r last:border-r-0"
                  style={{
                    height: HEADER_H,
                    borderColor: "var(--d-border)",
                    color: rc.hex,
                    minWidth: 120,
                  }}
                >
                  Комната {room === 2812 ? "812 (2к)" : room === 3812 ? "812 (3к)" : room}
                </div>
              );
            })}
          </div>

          {/* ── Body row ────────────────────────────────────────────────────── */}
          <div style={{ display: "flex" }}>

            {/* Time column — sticky left */}
            <div
              style={{
                width: TIME_COL_W,
                minWidth: TIME_COL_W,
                flexShrink: 0,
                borderRight: "1px solid var(--d-border)",
                background: "var(--d-surface)",
                position: "sticky",
                left: 0,
                zIndex: 10,
              }}
            >
              <div style={{ position: "relative", height: totalH }}>
                {hours.map((h) => (
                  <div
                    key={h}
                    className="absolute w-full flex items-start justify-end pr-2 pt-0.5"
                    style={{
                      top: (h - startH) * PX_PER_HOUR,
                      height: PX_PER_HOUR,
                    }}
                  >
                    <span
                      className="text-xs tnum select-none"
                      style={{ color: "var(--d-text-muted)" }}
                    >
                      {displayHour(h)}
                    </span>
                  </div>
                ))}

                {/* Now-dot at the time-axis / columns junction.
                    left:"auto" overrides the CSS class default left:-4px so
                    right:-4 places the dot at the RIGHT edge (junction). */}
                {nowTop !== null && nowTop >= 0 && nowTop <= totalH && (
                  <div
                    className="now-dot"
                    style={{ position: "absolute", left: "auto", right: -4, top: nowTop - 3 }}
                  />
                )}
              </div>
            </div>

            {/* Room columns */}
            <div style={{ display: "flex", flex: 1 }}>
              {roomOrder.map((room) => {
                const list = byRoom.get(room) ?? [];
                const rc = getRoomColor(room);

                return (
                  <div
                    key={room}
                    className="flex-1 border-r last:border-r-0"
                    style={{
                      position: "relative",
                      height: totalH,
                      minWidth: 120,
                      borderColor: "var(--d-border)",
                    }}
                  >
                    {/* Hour grid lines */}
                    {hours.map((h) => (
                      <div
                        key={h}
                        className="absolute w-full border-t"
                        style={{
                          top: (h - startH) * PX_PER_HOUR,
                          borderColor: "var(--d-border)",
                          opacity: 0.6,
                        }}
                      />
                    ))}

                    {/* Current-time indicator */}
                    {nowTop !== null && nowTop >= 0 && nowTop <= totalH && (
                      <div
                        className="absolute left-0 right-0 pointer-events-none"
                        style={{ top: nowTop }}
                      >
                        <div className="now-line-h" />
                      </div>
                    )}

                    {/* Bookings */}
                    {list.map((b) => {
                      const s = new Date(b.start);
                      const e = new Date(b.end);

                      let sh = s.getHours() + s.getMinutes() / 60;
                      let eh = e.getHours() + e.getMinutes() / 60;

                      if (s.getDate() !== d.getDate()) sh += 24;
                      if (e.getDate() !== d.getDate() && e.getDate() !== s.getDate()) eh += 24;
                      else if (e.getDate() !== d.getDate()) eh += 24;

                      const top = Math.max(0, (sh - startH) * PX_PER_HOUR);
                      const height = Math.max(18, (eh - sh) * PX_PER_HOUR);
                      const tiny = height < 36;

                      return (
                        <div
                          key={b.id}
                          className="absolute left-1 right-1 cursor-pointer z-[2]"
                          style={{ top, height }}
                          onClick={() => onInspect(b)}
                        >
                          <div
                            className="w-full h-full rounded-lg px-2 py-1 overflow-hidden transition-all duration-200 hover:shadow-md"
                            style={{
                              background: rc.hex + "28",
                              borderLeft: `3px solid ${rc.hex}`,
                              boxShadow: `0 1px 4px ${rc.hex}22`,
                            }}
                          >
                            {!tiny ? (
                              <>
                                <div
                                  className="font-semibold text-xs leading-tight line-clamp-2"
                                  style={{ color: "var(--d-text)" }}
                                >
                                  {b.title}
                                </div>
                                <div className="text-[10px] mt-0.5" style={{ color: "var(--d-text-sec)" }}>
                                  {toHM(b.start)}–{toHM(b.end)}
                                </div>
                              </>
                            ) : (
                              <div
                                className="text-[10px] font-medium truncate"
                                style={{ color: "var(--d-text)" }}
                              >
                                {b.title}
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>

          </div>
        </div>
      </div>
    </section>
  );
}
