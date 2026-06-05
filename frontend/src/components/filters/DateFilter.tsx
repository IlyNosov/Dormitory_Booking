import React, { useMemo, useRef } from "react";
import { toYMD } from "../../utils/date";
import { cn } from "../../utils/cn";

const RU_WEEKDAY_SHORT = ["вс", "пн", "вт", "ср", "чт", "пт", "сб"];
const RU_MONTH_SHORT = [
  "янв","фев","мар","апр","май","июн",
  "июл","авг","сен","окт","ноя","дек",
];

// Sentinel value meaning "no date filter" (today + future)
export const ALL_DATES = "all";
// Sentinel value meaning "past bookings only" (days strictly before today)
export const PAST_DATES = "past";

interface Props {
  value: string;
  onChange: (v: string) => void;
  hideAllDates?: boolean;
}

export function DateFilter({ value, onChange, hideAllDates }: Props) {
  const stripRef = useRef<HTMLDivElement>(null);

  const days = useMemo(() => {
    const result: { ymd: string; label: string; sub: string; isToday: boolean }[] = [];
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    for (let i = 0; i < 30; i++) {
      const d = new Date(today);
      d.setDate(today.getDate() + i);
      const ymd = toYMD(d);
      result.push({
        ymd,
        label: i === 0 ? "Сегодня" : RU_WEEKDAY_SHORT[d.getDay()],
        sub: `${d.getDate()} ${RU_MONTH_SHORT[d.getMonth()]}`,
        isToday: i === 0,
      });
    }
    return result;
  }, []);

  return (
    <div className="relative flex items-center gap-1">
      {/* Left arrow — desktop only */}
      <button
        type="button"
        className="hidden sm:flex shrink-0 w-8 h-8 items-center justify-center rounded-full hover:bg-[var(--d-panel)] transition-colors"
        onClick={() => { stripRef.current?.scrollBy({ left: -120, behavior: "smooth" }); }}
        aria-label="Назад"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M15 18l-6-6 6-6"/></svg>
      </button>

      {/* Scrollable strip */}
      <div ref={stripRef} className="date-strip flex-1 pb-1">
        {/* "Past" pill */}
        {!hideAllDates && (
          <button
            type="button"
            onClick={() => onChange(PAST_DATES)}
            className="flex-shrink-0 flex flex-col items-center px-4 py-2 rounded-xl border transition-all duration-200 ease-out"
            style={value === PAST_DATES ? {
              background: "var(--d-primary)",
              borderColor: "var(--d-primary)",
              color: "white",
            } : {
              background: "var(--d-surface)",
              borderColor: "var(--d-border)",
              color: "var(--d-text)",
            }}
            title="Прошедшие брони"
          >
            <span className="text-xs font-medium" style={{ color: value === PAST_DATES ? "rgba(255,255,255,0.8)" : "var(--d-text-muted)" }}>
              🕘
            </span>
            <span className="text-sm font-semibold mt-0.5">
              история
            </span>
          </button>
        )}

        {/* "All dates" pill — hidden in table mode */}
        {!hideAllDates && (
          <button
            type="button"
            onClick={() => onChange(ALL_DATES)}
            className={cn(
              "flex-shrink-0 flex flex-col items-center px-4 py-2 rounded-xl border transition-all duration-200 ease-out",
              value === ALL_DATES
                ? "text-white shadow-sm"
                : "hover:border-[color:var(--d-border)]"
            )}
            style={value === ALL_DATES ? {
              background: "var(--d-primary)",
              borderColor: "var(--d-primary)",
              color: "white",
            } : {
              background: "var(--d-surface)",
              borderColor: "var(--d-border)",
              color: "var(--d-text)",
            }}
            title="Все даты"
          >
            <span className="text-xs font-medium" style={{ color: value === ALL_DATES ? "rgba(255,255,255,0.8)" : "var(--d-text-muted)" }}>
              все
            </span>
            <span className="text-sm font-semibold mt-0.5">
              даты
            </span>
          </button>
        )}

        {days.map((day) => {
          const active = day.ymd === value;
          return (
            <button
              key={day.ymd}
              type="button"
              onClick={() => onChange(day.ymd)}
              className={cn(
                "flex-shrink-0 flex flex-col items-center px-4 py-2 rounded-xl border transition-all duration-200 ease-out"
              )}
              style={active ? {
                background: "var(--d-primary)",
                borderColor: "var(--d-primary)",
                color: "white",
              } : {
                background: "var(--d-surface)",
                borderColor: "var(--d-border)",
                color: "var(--d-text)",
              }}
            >
              <span className="text-xs font-medium" style={{ color: active ? "rgba(255,255,255,0.8)" : "var(--d-text-muted)" }}>
                {day.label}
              </span>
              <span className="text-sm font-semibold mt-0.5">
                {day.sub}
              </span>
            </button>
          );
        })}
      </div>

      {/* Right arrow — desktop only */}
      <button
        type="button"
        className="hidden sm:flex shrink-0 w-8 h-8 items-center justify-center rounded-full hover:bg-[var(--d-panel)] transition-colors"
        onClick={() => { stripRef.current?.scrollBy({ left: 120, behavior: "smooth" }); }}
        aria-label="Вперёд"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 18l6-6-6-6"/></svg>
      </button>
    </div>
  );
}
