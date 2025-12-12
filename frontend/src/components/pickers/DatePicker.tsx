import React, { useEffect, useMemo, useState } from "react";
import { cn } from "../../utils/cn";
import { fromYMD, pad2, toYMD } from "../../utils/date";

const RU_MONTHS = ["январь","февраль","март","апрель","май","июнь","июль","август","сентябрь","октябрь","ноябрь","декабрь"];
const RU_WEEKDAYS = ["пн","вт","ср","чт","пт","сб","вс"];

function monthMatrix(y: number, m: number) {
    const f = new Date(y, m, 1);
    const pad = (f.getDay() + 6) % 7;          // Пн=0
    const start = new Date(y, m, 1 - pad);
    return Array.from({ length: 42 }, (_, i) => new Date(start.getFullYear(), start.getMonth(), start.getDate() + i));
}

export function DatePicker({
                               value,
                               onChange,
                           }: {
    value: string;
    onChange: (v: string) => void;
}) {
    const sel = useMemo(() => fromYMD(value), [value]);
    const [y, setY] = useState(sel.getFullYear());
    const [m, setM] = useState(sel.getMonth());

    useEffect(() => {
        const d = fromYMD(value);
        setY(d.getFullYear());
        setM(d.getMonth());
    }, [value]);

    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const grid = useMemo(() => monthMatrix(y, m), [y, m]);

    function prevMonth() {
        const d = new Date(y, m - 1, 1);
        setY(d.getFullYear());
        setM(d.getMonth());
    }

    function nextMonth() {
        const d = new Date(y, m + 1, 1);
        setY(d.getFullYear());
        setM(d.getMonth());
    }

    return (
        <div className="w-[280px]">
            <div className="flex items-center justify-between mb-2">
                <button className="icon-btn" onClick={prevMonth}>‹</button>
                <div className="text-sm font-medium capitalize">
                    {RU_MONTHS[m]} {y}
                </div>
                <button className="icon-btn" onClick={nextMonth}>›</button>
            </div>

            <div className="grid grid-cols-7 text-[11px] text-zinc-400 mb-1">
                {RU_WEEKDAYS.map((w) => (
                    <div key={w} className="h-6 flex items-center justify-center uppercase">{w}</div>
                ))}
            </div>

            <div className="grid grid-cols-7 gap-1">
                {grid.map((d, i) => {
                    const cur = d.getMonth() === m;
                    const isToday = +d === +today;
                    const isSel = toYMD(d) === value;

                    return (
                        <button
                            key={i}
                            onClick={() => onChange(toYMD(d))}
                            className={cn(
                                "h-8 rounded-lg text-sm tnum transition select-none",
                                isSel ? "bg-blue-600 text-white" : "hover:bg-zinc-800/60",
                                cur ? "text-zinc-100" : "text-zinc-500",
                                isToday && !isSel && "ring-1 ring-blue-400/60",
                            )}
                        >
                            {pad2(d.getDate())}
                        </button>
                    );
                })}
            </div>
        </div>
    );
}
