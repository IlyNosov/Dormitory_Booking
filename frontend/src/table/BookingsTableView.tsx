import React, { useMemo } from "react";
import { Menu } from "lucide-react";
import { cn } from "../utils/cn";
import { pad2 } from "../utils/date";
import { ROOM_COLORS } from "../utils/rooms";
import type { Bookings } from "../types/bookings";

const ROOM_ORDER = [21, 256, 132] as const;

function isWeekend(d: Date) {
    const wd = d.getDay();
    return wd === 0 || wd === 6;
}

function dayLimits(d: Date) {
    const wd = d.getDay();
    const fri = wd === 5;
    const sat = wd === 6;
    return { startH: 6, endH: fri || sat ? 25 : 23 };
}

function toHM(dt: string) {
    const d = new Date(dt);
    return d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

function NowLine({ dayKey, startH, pxPerHour, color }: { dayKey: string; startH: number; pxPerHour: number; color: string }) {
    const now = new Date();
    const d = new Date(dayKey);
    if (now.toDateString() !== d.toDateString()) return null;

    let h = now.getHours() + now.getMinutes() / 60;
    const top = (h - startH) * pxPerHour;

    return (
        <div className="absolute left-0 right-0 pointer-events-none" style={{ top }}>
            <div className="now-line smooth" style={{ ["--nl" as any]: color }} />
        </div>
    );
}

export function BookingsTableView({
                                      dayKey,
                                      bookings,
                                      onInspect,
                                  }: {
    dayKey: string;
    bookings: Bookings[];
    onInspect: (b: Bookings) => void;
}) {
    const d = useMemo(() => new Date(dayKey), [dayKey]);
    const { startH, endH } = useMemo(() => dayLimits(d), [d]);
    const H = 48;
    const totalH = (endH - startH) * H;

    const byRoom = useMemo(() => {
        const map = new Map<(typeof ROOM_ORDER)[number], Bookings[]>();
        for (const r of ROOM_ORDER) map.set(r, []);
        for (const b of bookings) map.get(b.room as any)?.push(b);
        for (const [k, v] of map) v.sort((a, b) => +new Date(a.start) - +new Date(b.start));
        return map;
    }, [bookings]);

    return (
        <section>
            <div className="rounded-2xl border border-zinc-200 dark:border-[color:var(--d-border)] bg-white/90 dark:bg-[color:var(--d-card)] backdrop-blur p-4">
                <div className="grid grid-cols-[80px_repeat(3,1fr)] gap-3 min-w-[760px]">
                    <div className="relative pt-3 pb-6" style={{ height: totalH }}>
                        {Array.from({ length: endH - startH + 1 }, (_, i) => startH + i).map((h) => (
                            <div key={h} className="relative h-12">
                                <div className="absolute left-0 top-2 text-xs text-zinc-500 dark:text-zinc-400">
                                    {pad2(h % 24)}:00
                                </div>
                                <div className="absolute inset-x-0 top-7 border-t border-zinc-200 dark:border-[color:var(--d-border)]/70" />
                            </div>
                        ))}
                    </div>


                    {/* колонки */}
                    {ROOM_ORDER.map((room) => {
                        const list = byRoom.get(room) ?? [];
                        const rc = ROOM_COLORS[room];
                        const color = rc.hex;

                        return (
                            <div
                                key={room}
                                className="relative rounded-xl ring-1 ring-zinc-200 dark:ring-[color:var(--d-border)] bg-white/60 dark:bg-[color:var(--d-panel)]"
                                style={{ height: totalH }}
                            >
                                <div className="absolute top-2 left-3 text-xs uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
                                    Комната {room}
                                </div>

                                <NowLine dayKey={dayKey} startH={startH} pxPerHour={H} color={color} />

                                {list.map((b) => {
                                    const s = new Date(b.start);
                                    const e = new Date(b.end);

                                    let sh = s.getHours() + s.getMinutes() / 60;
                                    let eh = e.getHours() + e.getMinutes() / 60;

                                    if (isWeekend(d)) {
                                        if (s.getDate() !== d.getDate()) sh += 24;
                                        if (e.getDate() !== d.getDate()) eh += 24;
                                    }

                                    const top = (sh - startH) * H;
                                    const height = Math.max(14, (eh - sh) * H);
                                    const tiny = height < 32;

                                    return (
                                        <div
                                            key={b.id}
                                            className="absolute left-2 right-2 group cursor-pointer"
                                            style={{ top: Math.max(0, top), height }}
                                            onClick={() => onInspect(b)}
                                            title={b.description ? "Нажмите, чтобы увидеть описание" : undefined}
                                        >
                                            <div
                                                className={cn(
                                                    "w-full h-full rounded-xl px-2 py-1 text-[12px] transition-all duration-300",
                                                    "backdrop-blur-[2px] overflow-hidden ring-1 hover:shadow-lg",
                                                    rc.ring,
                                                    rc.bg,
                                                    "text-zinc-900 dark:text-zinc-100"
                                                )}
                                            >
                                                <div className="flex items-center justify-between gap-2">
                                                    {!tiny ? (
                                                        <div className="font-semibold leading-tight line-clamp-2">{b.title}</div>
                                                    ) : (
                                                        <div className="w-16 h-2 rounded-full opacity-90" style={{ background: color }} />
                                                    )}
                                                    {b.description && <Menu className="h-4 w-4 opacity-70 shrink-0" />}
                                                </div>

                                                {!tiny && (
                                                    <div className="text-[11px] opacity-80">
                                                        {toHM(b.start)}–{toHM(b.end)} · {b.isPrivate ? "Частная" : "Публичная"}
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
        </section>
    );
}
