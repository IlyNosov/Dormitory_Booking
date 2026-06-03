export const pad2 = (n: number) => String(n).padStart(2, "0");

export const toYMD = (d: Date) =>
    `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;

export const fromYMD = (s: string) => {
    const [yRaw, mRaw, dRaw] = s.split("-").map(Number);
    const y = Number.isFinite(yRaw) ? yRaw : new Date().getFullYear();
    const m = Number.isFinite(mRaw) ? mRaw : 1;
    const d = Number.isFinite(dRaw) ? dRaw : 1;
    const dt = new Date(y, m - 1, d);
    dt.setHours(0, 0, 0, 0);
    return dt;
};

export const msToHuman = (ms: number) => {
    const m = Math.max(0, Math.round(ms / 60000));
    const h = Math.floor(m / 60);
    const mm = m % 60;
    return h ? `${h}ч ${mm}м` : `${mm}м`;
};

export const fmtRange = (sISO: string, eISO: string) => {
    const s = new Date(sISO);
    const e = new Date(eISO);
    const d = s.toLocaleDateString("ru-RU", { day: "2-digit", month: "long" });
    const st = s.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
    const et = e.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
    return `${d} · ${st}–${et}`;
};

export const RU_MONTHS = [
    "январь","февраль","март","апрель","май","июнь",
    "июль","август","сентябрь","октябрь","ноябрь","декабрь",
];
export const RU_WEEKDAYS = ["пн","вт","ср","чт","пт","сб","вс"];

export function monthMatrix(y: number, m: number) {
    const f = new Date(y, m, 1);
    const pad = (f.getDay() + 6) % 7;
    const start = new Date(y, m, 1 - pad);
    return Array.from({ length: 42 }, (_, i) =>
        new Date(start.getFullYear(), start.getMonth(), start.getDate() + i),
    );
}
