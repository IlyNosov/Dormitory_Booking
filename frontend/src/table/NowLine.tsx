import React, { useMemo } from "react";

function clamp(n: number, a: number, b: number) {
    return Math.max(a, Math.min(b, n));
}

export function NowLine({
                            dayKey,
                            startHour,
                            pxPerHour,
                            color,
                        }: {
    dayKey: string;
    startHour: number;
    pxPerHour: number;
    color: string;
}) {
    const top = useMemo(() => {
        const now = new Date();
        const ymd = now.toISOString().slice(0, 10);
        if (ymd !== dayKey) return null;

        const h = now.getHours() + now.getMinutes() / 60;
        const t = (h - startHour) * pxPerHour;
        return clamp(t, 0, 10_000);
    }, [dayKey, startHour, pxPerHour]);

    if (top == null) return null;

    return (
        <div
            className="absolute left-0 right-0 pointer-events-none"
            style={{ top }}
        >
            <div
                className="now-line smooth"
                style={{ ["--nl" as any]: color }}
            />
        </div>
    );
}
