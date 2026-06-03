import React, { useEffect, useMemo, useRef, useState } from "react";
import { cn } from "../../utils/cn";

function pad2(n: number) {
    return String(n).padStart(2, "0");
}

function clamp(n: number, a: number, b: number) {
    return Math.max(a, Math.min(b, n));
}

export function TimePicker({
                               label,
                               value,
                               onChange,
                           }: {
    label: string;
    dateStr: string;
    value: string;
    onChange: (hm: string) => void;
}) {
    const [hh, mm] = value.split(":").map((x) => parseInt(x, 10));
    const hours = useMemo(() => Array.from({ length: 24 }, (_, i) => i), []);
    const minutes = useMemo(() => Array.from({ length: 12 }, (_, i) => i * 5), []);

    return (
        <div className="flex flex-col gap-2">
            <div className="text-sm lbl">{label}</div>

            <div className="relative grid grid-cols-[1fr_auto_1fr] items-center gap-2 rounded-2xl bg-zinc-100/60 dark:bg-[color:var(--d-panel)] px-3 py-2 max-w-[420px]">
                <div className="pointer-events-none absolute left-2 right-2 top-1/2 -translate-y-1/2 h-9 rounded-xl bg-white/70 dark:bg-white/5 ring-1 ring-zinc-300/70 dark:ring-white/10" />

                <WheelCol2
                    items={hours}
                    selected={Number.isFinite(hh) ? hh : 0}
                    format={pad2}
                    onPick={(h) => onChange(`${pad2(h)}:${pad2(mm || 0)}`)}
                    ariaLabel={`${label}: часы`}
                />

                <div className="z-10 text-zinc-500 dark:text-zinc-400 font-medium">:</div>

                <WheelCol2
                    items={minutes}
                    selected={Number.isFinite(mm) ? mm : 0}
                    format={pad2}
                    onPick={(m) => onChange(`${pad2(hh || 0)}:${pad2(m)}`)}
                    ariaLabel={`${label}: минуты`}
                />
            </div>
        </div>
    );
}

function WheelCol2({
                       items,
                       selected,
                       onPick,
                       ariaLabel,
                       format = (v: number) => String(v),
                   }: {
    items: number[];
    selected: number;
    onPick: (v: number) => void;
    ariaLabel: string;
    format?: (v: number) => string;
}) {
    const itemH = 40;
    const visible = 5;
    const mid = Math.floor(visible / 2);

    const idxFromValue = (v: number) => {
        const i = items.indexOf(v);
        return i >= 0 ? i : 0;
    };

    const [idx, setIdx] = useState(() => idxFromValue(selected));
    const idxRef = useRef(idx);
    useEffect(() => {
        idxRef.current = idx;
    }, [idx]);

    useEffect(() => {
        const i = idxFromValue(selected);
        setIdx(i);
    }, [selected]);

    const commit = (next: number) => {
        const clamped = clamp(next, 0, items.length - 1);
        setIdx(clamped);
        onPick(items[clamped]);
    };
    const hostRef = useRef<HTMLDivElement | null>(null);

    const wheelLockUntil = useRef(0);

    useEffect(() => {
        const el = hostRef.current;
        if (!el) return;

        const onWheelNative = (e: WheelEvent) => {
            e.preventDefault();
            e.stopPropagation();

            const now = Date.now();
            if (now < wheelLockUntil.current) return;

            const dir = e.deltaY > 0 ? 1 : -1;
            commit(idxRef.current + dir);

            wheelLockUntil.current = now + 60;
        };

        el.addEventListener("wheel", onWheelNative, { passive: false });
        return () => el.removeEventListener("wheel", onWheelNative as any);
    }, [items]);

    const drag = useRef<{ down: boolean; y: number; acc: number }>({ down: false, y: 0, acc: 0 });

    const onPointerDown: React.PointerEventHandler<HTMLDivElement> = (e) => {
        drag.current = { down: true, y: e.clientY, acc: 0 };
        (e.currentTarget as HTMLDivElement).setPointerCapture(e.pointerId);
    };

    const onPointerMove: React.PointerEventHandler<HTMLDivElement> = (e) => {
        if (!drag.current.down) return;
        const dy = e.clientY - drag.current.y;
        drag.current.y = e.clientY;
        drag.current.acc += dy;

        while (drag.current.acc >= itemH) {
            drag.current.acc -= itemH;
            commit(idxRef.current - 1);
        }
        while (drag.current.acc <= -itemH) {
            drag.current.acc += itemH;
            commit(idxRef.current + 1);
        }
    };

    const onPointerUp: React.PointerEventHandler<HTMLDivElement> = () => {
        drag.current.down = false;
        drag.current.acc = 0;
    };

    const offsetY = (mid - idx) * itemH;

    return (
        <div
            ref={hostRef}
            aria-label={ariaLabel}
            className={cn("relative z-10 h-[176px] overflow-hidden select-none wheel-capture mask-soft")}
            onPointerDown={onPointerDown}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
            onPointerCancel={onPointerUp}
            style={{ touchAction: "none" }}
        >
            <div className="absolute left-0 right-0 top-1/2 -translate-y-1/2" style={{ height: visible * itemH }}>
                <div className="will-change-transform transition-transform duration-150 ease-out" style={{ transform: `translateY(${offsetY}px)` }}>
                    {items.map((v, i) => {
                        const active = i === idx;
                        return (
                            <div
                                key={v}
                                className={cn("h-10 flex items-center justify-center tabular-nums font-medium", active ? "text-zinc-100" : "text-zinc-500")}
                            >
                                {format(v)}
                            </div>
                        );
                    })}
                </div>
            </div>
            <button type="button" className="absolute inset-x-0 top-0 h-1/2 cursor-pointer" onClick={() => commit(idxRef.current - 1)} aria-label="На шаг вверх" style={{ background: "transparent" }} />
            <button type="button" className="absolute inset-x-0 bottom-0 h-1/2 cursor-pointer" onClick={() => commit(idxRef.current + 1)} aria-label="На шаг вниз" style={{ background: "transparent" }} />
        </div>
    );
}

