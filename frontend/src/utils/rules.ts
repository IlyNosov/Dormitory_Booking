export const isWeekend = (d: Date) => [0, 6].includes(d.getDay());
export const isFriOrSat = (d: Date) => d.getDay() === 5 || d.getDay() === 6;

export const isQuiet = (t: Date) => {
    const h = t.getHours();
    const wd = t.getDay();
    const base = h >= 23 || h < 6;
    if (!base) return false;
    if (h >= 23 && (wd === 5 || wd === 6)) return false;
    return !(h === 0 && (wd === 6 || wd === 0));

};

export function dayLimits(d: Date) {
    return { startH: 6, endH: isWeekend(d) ? 25 : 23 };
}
