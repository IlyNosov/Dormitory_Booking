import React, { useEffect, useMemo, useRef, useState } from "react";
import { AlertTriangle, Lock, Unlock, ChevronRight, Send } from "lucide-react";

import type { Booking, CreateBookingPayload, RoomFilter, ViewMode, VisFilter, RoomInfo } from "../types/bookings";
import { cn } from "../utils/cn";
import { fmtRange, fromYMD, msToHuman, pad2, toYMD } from "../utils/date";
import { isQuiet } from "../utils/rules";
import { normalizeApiError } from "../utils/errors";

import { AppHeader } from "./AppHeader";
import { DateFilter, ALL_DATES, PAST_DATES } from "../components/filters/DateFilter";
import { TimePicker } from "../components/time/TimePicker";
import { BookingCard, displayRoom } from "../components/BookingCard";
import { Modal } from "../components/ui/Modal";
import { DatePicker } from "../components/pickers/DatePicker";
import { RulesModal } from "../components/ui/RulesModal";
import { BookingsTableView } from "../table/BookingsTableView";
import { AuthModal } from "../auth/AuthModal";
import { AdminPanel } from "../admin/AdminPanel";
import { RoomsPage } from "../pages/RoomsPage";
import { useAuth } from "../auth/useAuth";
import * as api from "../api/bookings";
import { fetchRooms } from "../api/rooms";

// ─── constants ────────────────────────────────────────────────────────────────
const RU_WEEKDAY_SHORT = ["вс", "пн", "вт", "ср", "чт", "пт", "сб"];
const RU_MONTH_SHORT = ["янв","фев","мар","апр","май","июн","июл","авг","сен","окт","ноя","дек"];

// ─── BookingWizard ─────────────────────────────────────────────────────────────
interface WizardForm {
  date: string;
  startH: number;
  startM: number;
  endH: number;
  endM: number;
  room: number | null;
  title: string;
  description: string;
  isPrivate: boolean;
  telegramId: string;
}

// Step label component
function StepLabel({ n, label }: { n: number; label: string }) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <div
        className="flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold text-white shrink-0"
        style={{ background: "var(--d-primary)" }}
      >
        {n}
      </div>
      <span className="text-sm font-semibold" style={{ color: "var(--d-text)" }}>{label}</span>
    </div>
  );
}

// 14-day date strip for wizard
function WizardDateStrip({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const [showCalendar, setShowCalendar] = useState(false);
  const stripRef = useRef<HTMLDivElement>(null);
  const days = useMemo(() => {
    const result: { ymd: string; label: string; sub: string }[] = [];
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    for (let i = 0; i < 14; i++) {
      const d = new Date(today);
      d.setDate(today.getDate() + i);
      const ymd = toYMD(d);
      result.push({
        ymd,
        label: i === 0 ? "Сег" : RU_WEEKDAY_SHORT[d.getDay()],
        sub: `${d.getDate()} ${RU_MONTH_SHORT[d.getMonth()]}`,
      });
    }
    return result;
  }, []);

  const isInStrip = days.some(d => d.ymd === value);

  return (
    <div className="space-y-3">
      <div className="relative flex items-center gap-1">
        {/* Left arrow — desktop only */}
        <button
          type="button"
          className="hidden sm:flex shrink-0 w-8 h-8 items-center justify-center rounded-full hover:bg-[var(--d-panel)] transition-colors"
          onClick={() => stripRef.current?.scrollBy({ left: -120, behavior: "smooth" })}
          aria-label="Назад"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M15 18l-6-6 6-6"/></svg>
        </button>

      <div ref={stripRef} className="date-strip flex-1 pb-1">
        {days.map((day) => {
          const active = day.ymd === value;
          return (
            <button
              key={day.ymd}
              type="button"
              onClick={() => { onChange(day.ymd); setShowCalendar(false); }}
              className="flex-shrink-0 flex flex-col items-center px-3 py-1.5 rounded-xl border transition-all duration-200 ease-out"
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
              <span className="text-[10px] font-medium" style={{ color: active ? "rgba(255,255,255,0.8)" : "var(--d-text-muted)" }}>
                {day.label}
              </span>
              <span className="text-xs font-semibold mt-0.5">
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
          onClick={() => stripRef.current?.scrollBy({ left: 120, behavior: "smooth" })}
          aria-label="Вперёд"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 18l6-6-6-6"/></svg>
        </button>
      </div>

      {/* "Other date" toggle */}
      <button
        type="button"
        className="text-xs underline underline-offset-2"
        style={{ color: "var(--d-text-muted)" }}
        onClick={() => setShowCalendar(v => !v)}
      >
        {showCalendar ? "Скрыть календарь" : "Другая дата"}
        {!isInStrip && value ? ` (выбрано: ${value})` : ""}
      </button>

      {showCalendar && (
        <div className="overflow-x-auto flex justify-center">
          <div className="mx-auto max-w-fit">
            <DatePicker value={value} onChange={(v) => { onChange(v); setShowCalendar(false); }} />
          </div>
        </div>
      )}
    </div>
  );
}

interface BookingWizardProps {
  activeRooms: RoomInfo[];
  onSubmit: (form: WizardForm) => Promise<void>;
  onCancel: () => void;
  formLoading: boolean;
  formError: string;
  initialBooking?: Booking;
}

function BookingWizard({ activeRooms, onSubmit, onCancel, formLoading, formError, initialBooking }: BookingWizardProps) {
  const topRef = useRef<HTMLDivElement>(null);
  const now = new Date();

  const defaultForm: WizardForm = initialBooking ? (() => {
    const s = new Date(initialBooking.start);
    const e = new Date(initialBooking.end);
    return {
      date: toYMD(s),
      startH: s.getHours(),
      startM: s.getMinutes(),
      endH: e.getHours(),
      endM: e.getMinutes(),
      room: initialBooking.room,
      title: initialBooking.title,
      description: initialBooking.description ?? "",
      isPrivate: initialBooking.isPrivate,
      telegramId: initialBooking.telegramId ?? "",
    };
  })() : {
    date: toYMD(now),
    startH: now.getHours() + 1,
    startM: 0,
    endH: now.getHours() + 2,
    endM: 0,
    room: activeRooms[0]?.number ?? null,
    title: "",
    description: "",
    isPrivate: true,
    telegramId: localStorage.getItem("savedTelegramId") ?? "",
  };

  const [form, setForm] = useState<WizardForm>(defaultForm);
  const [extraOpen, setExtraOpen] = useState(false);

  // Scroll to top of modal when error appears
  useEffect(() => {
    if (!formError) return;
    let el = topRef.current?.parentElement;
    while (el) {
      const s = getComputedStyle(el);
      if (s.overflowY === "auto" || s.overflowY === "scroll") {
        el.scrollTo({ top: 0, behavior: "smooth" });
        return;
      }
      el = el.parentElement;
    }
  }, [formError]);

  // Duration
  const startMins = form.startH * 60 + form.startM;
  const endMins = form.endH * 60 + form.endM;
  const durationMs = (endMins - startMins) * 60000;
  const durationValid = endMins > startMins;

  // Format selected date for summary
  const selectedDate = form.date ? (() => {
    const d = fromYMD(form.date);
    return d.toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit", year: "numeric" });
  })() : null;

  const selectedRoom = form.room !== null
    ? activeRooms.find(r => r.number === form.room)
    : null;

  return (
    <div className="space-y-5 pt-2" ref={topRef}>
      <div className="text-lg font-semibold" style={{ color: "var(--d-primary-h)" }}>
        {initialBooking ? "Редактировать бронирование" : "Новое бронирование"}
      </div>

      {isQuiet(new Date()) && (
        <div className="banner-warn">
          <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
          <span>Сейчас тихий час. Убедитесь, что выбранное время разрешено.</span>
        </div>
      )}

      {formError && (
        <div className="banner-error">
          <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
          <span>{formError}</span>
        </div>
      )}

      {/* ── Step 1: Room ─────────────────────────────────────────────────── */}
      <div className="rounded-2xl border p-4 space-y-3" style={{ background: "var(--d-panel)", borderColor: "var(--d-border)" }}>
        <StepLabel n={1} label="Выберите комнату" />
        <div className="flex flex-wrap gap-2">
          {activeRooms.map(r => {
            const active = form.room === r.number;
            return (
              <button
                key={r.number}
                type="button"
                onClick={() => setForm(f => ({ ...f, room: r.number }))}
                className="px-3 py-2 rounded-xl border text-sm font-medium transition-all duration-200"
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
                {r.name || `Комната ${r.number}`}
              </button>
            );
          })}
        </div>
      </div>

      {/* ── Step 2: Date ─────────────────────────────────────────────────── */}
      <div className="rounded-2xl border p-4 space-y-3" style={{ background: "var(--d-panel)", borderColor: "var(--d-border)" }}>
        <StepLabel n={2} label="Выберите дату" />
        <WizardDateStrip value={form.date} onChange={(v) => setForm(f => ({ ...f, date: v }))} />
      </div>

      {/* ── Step 3: Time ─────────────────────────────────────────────────── */}
      <div className="rounded-2xl border p-4 space-y-3" style={{ background: "var(--d-panel)", borderColor: "var(--d-border)" }}>
        <StepLabel n={3} label="Выберите время" />
        <div className="grid grid-cols-2 gap-4">
          <TimePicker
            label="Начало"
            dateStr={form.date}
            value={`${pad2(form.startH)}:${pad2(form.startM)}`}
            onChange={(hm) => {
              const [h, m] = hm.split(":").map(Number);
              setForm(f => ({ ...f, startH: h, startM: m }));
            }}
          />
          <TimePicker
            label="Конец"
            dateStr={form.date}
            value={`${pad2(form.endH)}:${pad2(form.endM)}`}
            onChange={(hm) => {
              const [h, m] = hm.split(":").map(Number);
              setForm(f => ({ ...f, endH: h, endM: m }));
            }}
          />
        </div>
        {/* Duration indicator */}
        <div className="text-sm" style={{ color: durationValid ? "var(--d-text-sec)" : "var(--d-error)" }}>
          {durationValid
            ? `⏱ ${msToHuman(durationMs)}`
            : "⚠ Конец раньше начала"}
        </div>
      </div>

      {/* ── Step 4: Event details ─────────────────────────────────────────── */}
      <div className="rounded-2xl border p-4 space-y-3" style={{ background: "var(--d-panel)", borderColor: "var(--d-border)" }}>
        <StepLabel n={4} label="Опишите мероприятие" />

        <label className="flex flex-col gap-1 text-sm">
          <div className="flex items-center justify-between">
            <span className="lbl">Название *</span>
            <span className="text-xs" style={{ color: form.title.length >= 60 ? "var(--d-error)" : "var(--d-text-muted)" }}>
              {form.title.length}/80
            </span>
          </div>
          <input
            className="field"
            maxLength={80}
            value={form.title}
            onChange={(e) => setForm(f => ({ ...f, title: e.target.value }))}
            placeholder="Кино, настолки, репетиция…"
          />
        </label>

        <label className="flex flex-col gap-1 text-sm">
          <div className="flex items-center justify-between">
            <span className="lbl">Telegram для связи *</span>
            <span className="text-xs" style={{ color: form.telegramId.length >= 28 ? "var(--d-error)" : "var(--d-text-muted)" }}>
              {form.telegramId.length}/32
            </span>
          </div>
          <div className="relative">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--d-text-muted)]">@</span>
            <input
              className="field pl-7"
              maxLength={32}
              placeholder="username"
              value={form.telegramId}
              onChange={e => setForm(f => ({ ...f, telegramId: e.target.value.replace(/^@/, '') }))}
            />
          </div>
        </label>

        {/* Private/public toggle — segment control */}
        <div>
          <span className="lbl block mb-1.5">Тип</span>
          <div
            className="inline-flex rounded-xl p-1 gap-1"
            style={{ background: "var(--d-panel)" }}
          >
            <button
              type="button"
              onClick={() => setForm(f => ({ ...f, isPrivate: false }))}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-200"
              style={!form.isPrivate ? {
                background: "var(--d-primary)",
                color: "white",
              } : {
                background: "transparent",
                color: "var(--d-text-muted)",
              }}
            >
              <Unlock className="h-3.5 w-3.5" />
              Публичное
            </button>
            <button
              type="button"
              onClick={() => setForm(f => ({ ...f, isPrivate: true }))}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-200"
              style={form.isPrivate ? {
                background: "var(--d-primary)",
                color: "white",
              } : {
                background: "transparent",
                color: "var(--d-text-muted)",
              }}
            >
              <Lock className="h-3.5 w-3.5" />
              Частное
            </button>
          </div>
        </div>

        {/* Collapsible "Дополнительно" — CSS grid accordion */}
        <div>
          <button
            type="button"
            className="flex items-center gap-1 text-xs font-medium transition-colors"
            style={{ color: "var(--d-text-muted)" }}
            onClick={() => setExtraOpen(v => !v)}
          >
            <ChevronRight
              className="h-3.5 w-3.5 transition-transform duration-200"
              style={{ transform: extraOpen ? "rotate(90deg)" : "none" }}
            />
            Дополнительно
          </button>
          <div className={`accordion-body${extraOpen ? " open" : ""}`}>
            <div>
              <label className="flex flex-col gap-1 text-sm pt-2">
                <span className="lbl">Описание</span>
                <div className="relative">
                  <textarea
                    className="field resize-none"
                    rows={3}
                    maxLength={300}
                    value={form.description}
                    onChange={(e) => setForm(f => ({ ...f, description: e.target.value }))}
                    placeholder="Дополнительная информация (необязательно)"
                  />
                  <span
                    className="absolute bottom-2 right-2 text-xs pointer-events-none"
                    style={{ color: "var(--d-text-muted)" }}
                  >
                    {form.description.length}/300
                  </span>
                </div>
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* ── Summary block ─────────────────────────────────────────────────── */}
      <div
        className="rounded-xl p-4 space-y-1.5 text-sm border-l-4"
        style={{
          background: "var(--d-panel)",
          borderLeftColor: "var(--d-primary)",
          border: "1px solid var(--d-border)",
          borderLeft: "4px solid var(--d-primary)",
        }}
      >
        <div className="font-semibold text-xs uppercase tracking-wide mb-2" style={{ color: "var(--d-text-muted)" }}>
          Ваше бронирование
        </div>
        <div className="flex gap-2">
          <span style={{ color: "var(--d-text-muted)" }}>Комната:</span>
          <span style={{ color: "var(--d-text)" }}>
            {selectedRoom ? (selectedRoom.name || `Комната ${selectedRoom.number}`) : <em style={{ color: "var(--d-text-muted)" }}>не выбрано</em>}
          </span>
        </div>
        <div className="flex gap-2">
          <span style={{ color: "var(--d-text-muted)" }}>Дата:</span>
          <span style={{ color: "var(--d-text)" }}>
            {selectedDate ?? <em style={{ color: "var(--d-text-muted)" }}>не выбрано</em>}
          </span>
        </div>
        <div className="flex gap-2">
          <span style={{ color: "var(--d-text-muted)" }}>Время:</span>
          <span style={{ color: "var(--d-text)" }}>
            {durationValid
              ? `${pad2(form.startH)}:${pad2(form.startM)} – ${pad2(form.endH)}:${pad2(form.endM)} (${msToHuman(durationMs)})`
              : <em style={{ color: "var(--d-text-muted)" }}>не выбрано</em>}
          </span>
        </div>
        {form.title && (
          <div className="flex gap-2 min-w-0 overflow-hidden">
            <span className="shrink-0" style={{ color: "var(--d-text-muted)" }}>Название:</span>
            <span className="truncate min-w-0" style={{ color: "var(--d-text)" }}>{form.title}</span>
          </div>
        )}
        {form.telegramId && (
          <div className="flex gap-2 min-w-0 overflow-hidden">
            <span className="shrink-0" style={{ color: "var(--d-text-muted)" }}>Telegram:</span>
            <span className="truncate min-w-0" style={{ color: "var(--d-text)" }}>@{form.telegramId}</span>
          </div>
        )}
      </div>

      {/* ── Actions ───────────────────────────────────────────────────────── */}
      <div className="flex flex-col gap-2 pt-1">
        <div className="flex gap-2">
          <button
            className="btn btn-primary"
            onClick={() => onSubmit(form)}
            disabled={formLoading}
          >
            {formLoading
              ? (initialBooking ? "Сохранение…" : "Создание…")
              : (initialBooking ? "Сохранить" : "Забронировать")}
          </button>
          <button className="btn" onClick={onCancel}>
            Отмена
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Main App ──────────────────────────────────────────────────────────────────
export function App() {
  const auth = useAuth();
  const [authOpen, setAuthOpen] = useState(false);
  const [adminOpen, setAdminOpen] = useState(false);
  const [rulesOpen, setRulesOpen] = useState(false);
  const [myBookings, setMyBookings] = useState(false);

  const [bookings, setBookings] = useState<Booking[]>([]);
  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [loading, setLoading] = useState(true);

  const [roomFilter, setRoomFilter] = useState<RoomFilter>("all");
  const [visFilter, setVisFilter] = useState<VisFilter>("all");
  const [view, setView] = useState<ViewMode>("cards");
  const [filterDate, setFilterDate] = useState<string>(() => toYMD(new Date()));

  // View transition state
  const [viewTransition, setViewTransition] = useState(false);

  const [addOpen, setAddOpen] = useState(false);
  const [inspectBooking, setInspectBooking] = useState<Booking | null>(null);
  const [editBooking, setEditBooking] = useState<Booking | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [page, setPage] = useState<"main" | "rooms">("main");

  const [formError, setFormError] = useState("");
  const [formLoading, setFormLoading] = useState(false);

  // silent=true: update data without showing the skeleton loading state.
  async function loadBookings(silent = false) {
    if (!silent) setLoading(true);
    try { setBookings(await api.fetchBookings()); }
    finally { if (!silent) setLoading(false); }
  }

  async function loadRooms() {
    if (auth.isAuth) {
      try {
        const data = await fetchRooms();
        setRooms(data);
      } catch (e: unknown) {
        console.error("[App] loadRooms:", e);
      }
    }
  }

  useEffect(() => { loadBookings(); }, []);

  // Reload bookings on auth change (login/logout) so canManage is fresh
  useEffect(() => {
    if (auth.user !== undefined) loadBookings();
  }, [auth.isAuth]);

  useEffect(() => { loadRooms(); }, [auth.isAuth]);

  // SSE: instant updates whenever any client mutates bookings.
  // Falls back to a slow 5-minute poll if the stream is unavailable.
  useEffect(() => {
    const source = new EventSource("/api/bookings/events");
    source.addEventListener("update", () => loadBookings(true));
    // 5-minute silent fallback poll
    const fallback = setInterval(() => loadBookings(true), 5 * 60_000);
    // Refresh on tab re-focus (handles long backgrounded tabs)
    const onVisible = () => { if (document.visibilityState === "visible") loadBookings(true); };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      source.close();
      clearInterval(fallback);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, []);

  const ROOM_ORDER = [21, 256, 132, 2812, 3812];
  const activeRooms = useMemo(() =>
    rooms.filter(r => r.isActive).sort((a, b) => {
      const ai = ROOM_ORDER.indexOf(a.number);
      const bi = ROOM_ORDER.indexOf(b.number);
      if (ai === -1 && bi === -1) return a.number - b.number;
      if (ai === -1) return 1;
      if (bi === -1) return -1;
      return ai - bi;
    }),
  [rooms]);

  const todayStart = useMemo(() => {
    const d = new Date(); d.setHours(0, 0, 0, 0); return d;
  }, []);

  const filtered = useMemo(() => {
    const myEmail = auth.user?.email;
    return bookings.filter((b) => {
      const bStart = new Date(b.start);
      if (filterDate === PAST_DATES) {
        // Only bookings from days strictly before today
        if (bStart >= todayStart) return false;
      } else if (filterDate === ALL_DATES) {
        // Hide days strictly before today; today's bookings always visible
        if (bStart < todayStart) return false;
      } else {
        // Specific date
        const dayStart = fromYMD(filterDate);
        const dayEnd = new Date(dayStart);
        dayEnd.setDate(dayEnd.getDate() + 1);
        if (bStart < dayStart || bStart >= dayEnd) return false;
      }
      if (roomFilter !== "all" && b.room !== roomFilter) return false;
      if (visFilter === "public" && b.isPrivate) return false;
      if (visFilter === "private" && !b.isPrivate) return false;
      if (myBookings && b.userEmail !== myEmail) return false;
      return true;
    });
  }, [bookings, filterDate, roomFilter, visFilter, myBookings, auth.user, todayStart]);

  function handleViewChange(v: ViewMode) {
    setViewTransition(true);
    if (v === "table" && (filterDate === ALL_DATES || filterDate === PAST_DATES)) {
      setFilterDate(toYMD(new Date()));
    }
    setTimeout(() => { setView(v); setViewTransition(false); }, 150);
  }

  async function handleCreate(form: {
    date: string; startH: number; startM: number; endH: number; endM: number;
    room: number | null; title: string; description: string; isPrivate: boolean; telegramId: string;
  }) {
    setFormError("");

    if (form.room === null) { setFormError("Выберите комнату"); return; }
    if (!form.title.trim()) { setFormError("Укажите название"); return; }
    if (!form.telegramId.trim()) { setFormError("Укажите Telegram для связи"); return; }

    const day = fromYMD(form.date);
    const start = new Date(day);
    start.setHours(form.startH, form.startM, 0, 0);
    const end = new Date(day);
    end.setHours(form.endH, form.endM, 0, 0);

    if (end <= start) { setFormError("Конец должен быть позже начала"); return; }

    setFormLoading(true);
    const selectedRoom = activeRooms.find(r => r.number === form.room) ?? activeRooms[0];
    if (!selectedRoom) { setFormError("Нет доступных комнат для бронирования"); setFormLoading(false); return; }

    try {
      const payload: CreateBookingPayload = {
        start: start.toISOString(),
        end: end.toISOString(),
        room: selectedRoom.number,
        title: form.title.trim(),
        description: form.description.trim() || undefined,
        isPrivate: form.isPrivate,
        telegram_id: form.telegramId.trim() || undefined,
      };
      const created = await api.createBooking(payload);
      if (form.telegramId.trim()) localStorage.setItem("savedTelegramId", form.telegramId.trim());
      // Optimistically add to state — SSE will sync the rest seamlessly.
      setBookings(prev => [...prev, created]);
      setAddOpen(false);
      setFormError("");
    } catch (e: unknown) {
      const msg = normalizeApiError(e);
      if (msg.includes("409") || msg.toLowerCase().includes("conflict") || msg.toLowerCase().includes("overlap")) {
        setFormError("Такое время уже занято");
      } else {
        setFormError(msg);
      }
      console.error("[App] createBooking:", e);
    } finally {
      setFormLoading(false);
    }
  }

  function requestDelete(id: string) {
    setConfirmDeleteId(id);
  }

  async function handleDelete(id: string) {
    setConfirmDeleteId(null);
    // Optimistic: remove immediately, revert on error.
    setBookings(bs => bs.filter(b => b.id !== id));
    if (inspectBooking?.id === id) setInspectBooking(null);
    try {
      await api.deleteBooking(id);
    } catch (e: unknown) {
      loadBookings(true);
      console.error("[App] deleteBooking:", e);
    }
  }

  async function handleUpdate(form: {
    date: string; startH: number; startM: number; endH: number; endM: number;
    room: number | null; title: string; description: string; isPrivate: boolean; telegramId: string;
  }) {
    if (!editBooking) return;
    setFormError("");

    if (form.room === null) { setFormError("Выберите комнату"); return; }
    if (!form.title.trim()) { setFormError("Укажите название"); return; }

    const day = fromYMD(form.date);
    const start = new Date(day);
    start.setHours(form.startH, form.startM, 0, 0);
    const end = new Date(day);
    end.setHours(form.endH, form.endM, 0, 0);
    if (end <= start) { setFormError("Конец должен быть позже начала"); return; }

    setFormLoading(true);
    try {
      const updated = await api.updateBooking(editBooking.id, {
        start: start.toISOString(),
        end: end.toISOString(),
        room: form.room,
        title: form.title.trim(),
        description: form.description.trim() || undefined,
        isPrivate: form.isPrivate,
        telegram_id: form.telegramId.trim() || undefined,
      });
      if (form.telegramId.trim()) localStorage.setItem("savedTelegramId", form.telegramId.trim());
      setBookings(bs => bs.map(b => b.id === updated.id ? updated : b));
      setEditBooking(null);
      setFormError("");
    } catch (e: unknown) {
      const msg = normalizeApiError(e);
      if (msg.includes("409") || msg.toLowerCase().includes("conflict") || msg.toLowerCase().includes("overlap")) {
        setFormError("Такое время уже занято");
      } else {
        setFormError(msg);
      }
      console.error("[App] updateBooking:", e);
    } finally {
      setFormLoading(false);
    }
  }

  if (page === "rooms") {
    return (
      <div className="page-slide-in">
      <RoomsPage
        onBack={() => setPage("main")}
        onBook={() => {
          setPage("main");
          if (!auth.isAuth) { setAuthOpen(true); return; }
          setAddOpen(true);
        }}
      />
      </div>
    );
  }

  const contentKey = `${filterDate}-${String(roomFilter)}-${visFilter}-${myBookings}`;

  return (
    <div className="min-h-screen" style={{ background: "var(--d-bg)" }}>
      <AppHeader
        view={view}
        user={auth.user}
        myBookings={myBookings}
        onToggleAdd={() => {
          if (!auth.isAuth) { setAuthOpen(true); return; }
          setAddOpen(true);
        }}
        onToggleView={handleViewChange}
        onToggleMyBookings={() => setMyBookings(m => !m)}
        onLoginClick={() => setAuthOpen(true)}
        onAdminClick={() => setAdminOpen(true)}
        onLogout={auth.logout}
        onRulesClick={() => setRulesOpen(true)}
        onRoomsClick={() => setPage("rooms")}
      />

      <main className="mx-auto max-w-6xl px-4 py-6 space-y-6">
        {/* Filters */}
        <div className="flex flex-col gap-3">
          <DateFilter
            value={filterDate}
            onChange={setFilterDate}
            hideAllDates={view === "table"}
          />

          {/* Room filter buttons — hidden in table mode */}
          {activeRooms.length > 0 && view !== "table" && (
            <div className="flex flex-wrap gap-2 items-center">
              <span className="lbl mr-1">Комната:</span>
              <button
                className={cn("btn-filter", roomFilter === "all" && "active")}
                onClick={() => setRoomFilter("all")}
              >
                Все
              </button>
              {activeRooms.map(r => (
                <button
                  key={r.number}
                  className={cn("btn-filter", roomFilter === r.number && "active")}
                  onClick={() => setRoomFilter(r.number)}
                >
                  {r.name || `Комната ${r.number}`}
                </button>
              ))}
            </div>
          )}

          <div className="flex flex-wrap gap-2 items-center">
            <span className="lbl mr-1">Тип:</span>
            {(["all", "public", "private"] as VisFilter[]).map(opt => (
              <button
                key={opt}
                className={cn("btn-filter", visFilter === opt && "active")}
                onClick={() => setVisFilter(opt)}
              >
                {opt === "all" ? "Все" : opt === "public" ? "Публичные" : "Частные"}
              </button>
            ))}
          </div>
        </div>

        {/* Content with transition */}
        <div
          style={{
            transition: "opacity 0.18s ease-out, transform 0.18s ease-out",
            opacity: viewTransition ? 0 : 1,
            transform: viewTransition ? "translateY(6px) scale(0.99)" : "none",
          }}
        >
          {view === "table" && (
            <BookingsTableView
              dayKey={filterDate}
              bookings={filtered}
              onInspect={setInspectBooking}
            />
          )}

          {view === "cards" && (() => {
            if (loading) {
              return (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="h-36 rounded-2xl animate-pulse" style={{ background: "var(--d-border)" }} />
                  ))}
                </div>
              );
            }
            if (filtered.length === 0) {
              return (
                <div className="text-center py-16 text-sm" style={{ color: "var(--d-text-muted)" }}>
                  Нет бронирований на выбранный день
                </div>
              );
            }
            if (filterDate !== ALL_DATES && filterDate !== PAST_DATES) {
              return (
                <div key={contentKey} className="content-fade-in grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filtered.map((b, i) => (
                    <div
                      key={b.id}
                      className="card-enter"
                      style={{ animationDelay: `${Math.min(i * 35, 210)}ms` }}
                    >
                      <BookingCard
                        booking={b}
                        onDelete={b.canManage ? () => requestDelete(b.id) : undefined}
                        onInspect={() => setInspectBooking(b)}
                      />
                    </div>
                  ))}
                </div>
              );
            }
            // "Все даты" / "История" — grouped by date with headers
            const groups = new Map<string, typeof filtered>();
            for (const b of filtered) {
              const key = toYMD(new Date(b.start));
              if (!groups.has(key)) groups.set(key, []);
              groups.get(key)!.push(b);
            }
            // History shows newest first; future shows oldest first
            const sortedKeys = [...groups.keys()].sort(
              filterDate === PAST_DATES ? (a, b) => b.localeCompare(a) : undefined
            );
            const todayStr = toYMD(new Date());
            const tomorrowDate = new Date();
            tomorrowDate.setDate(tomorrowDate.getDate() + 1);
            const tomorrowStr = toYMD(tomorrowDate);
            let cardIdx = 0;
            return (
              <div key={contentKey} className="content-fade-in space-y-6">
                {sortedKeys.map((key, gi) => {
                  const dateObj = new Date(key + "T00:00:00");
                  const dayMonth = dateObj.toLocaleDateString("ru-RU", { day: "numeric", month: "long" });
                  const label = key === todayStr
                    ? `Сегодня, ${dayMonth}`
                    : key === tomorrowStr
                    ? `Завтра, ${dayMonth}`
                    : dayMonth;
                  return (
                    <div
                      key={key}
                      className="space-y-3 card-enter"
                      style={{ animationDelay: `${gi * 55}ms` }}
                    >
                      <div className="text-sm font-semibold capitalize" style={{ color: "var(--d-text-sec)" }}>
                        {label}
                      </div>
                      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                        {groups.get(key)!.map(b => {
                          const delay = Math.min(cardIdx++ * 30, 240);
                          return (
                            <div key={b.id} className="card-enter" style={{ animationDelay: `${delay}ms` }}>
                              <BookingCard
                                booking={b}
                                onDelete={b.canManage ? () => requestDelete(b.id) : undefined}
                                onInspect={() => setInspectBooking(b)}
                              />
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            );
          })()}
        </div>
      </main>

      {/* Auth modal */}
      <AuthModal
        open={authOpen}
        onClose={() => setAuthOpen(false)}
        onSuccess={() => {
          auth.refresh();
          setAuthOpen(false);
        }}
      />

      {/* Admin panel */}
      <AdminPanel
        open={adminOpen}
        onClose={() => setAdminOpen(false)}
        isSuperAdmin={auth.isSuperAdmin}
        onForcePushSuccess={() => { setAdminOpen(false); loadBookings(); }}
      />

      {/* Rules modal */}
      <RulesModal open={rulesOpen} onClose={() => setRulesOpen(false)} />

      {/* Create booking modal — 4-step wizard */}
      <Modal open={addOpen} onClose={() => { setAddOpen(false); setFormError(""); }}>
        <BookingWizard
          activeRooms={activeRooms}
          onSubmit={handleCreate}
          onCancel={() => { setAddOpen(false); setFormError(""); }}
          formLoading={formLoading}
          formError={formError}
        />
      </Modal>

      {/* Inspect booking modal */}
      {inspectBooking && (
        <Modal open={!!inspectBooking} onClose={() => setInspectBooking(null)}>
          <div className="space-y-3 min-w-0 overflow-hidden">
            <div
              className="text-lg font-semibold leading-snug break-words"
              style={{ color: "var(--d-primary-h)" }}
            >
              {inspectBooking.title.slice(0, 120)}{inspectBooking.title.length > 120 ? "…" : ""}
            </div>
            <div className="text-sm" style={{ color: "var(--d-text-sec)" }}>
              Комната {displayRoom(inspectBooking.room)} ·{" "}
              {fmtRange(inspectBooking.start, inspectBooking.end)}
            </div>
            {inspectBooking.isPrivate && (
              <div className="flex items-center gap-1.5 text-sm" style={{ color: "var(--d-secondary)" }}>
                <Lock className="h-3.5 w-3.5" /> Частное
              </div>
            )}
            {inspectBooking.description && (
              <p className="text-sm leading-relaxed break-words" style={{ color: "var(--d-text)" }}>{inspectBooking.description}</p>
            )}
            {inspectBooking.telegramId && (
              <div className="flex items-center gap-1.5 text-sm overflow-hidden" style={{ color: "var(--d-text-muted)" }}>
                <Send className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate min-w-0">
                  @{inspectBooking.telegramId.slice(0, 40)}{inspectBooking.telegramId.length > 40 ? "…" : ""}
                </span>
              </div>
            )}
            {inspectBooking.userEmail && (
              <div className="text-xs truncate" style={{ color: "var(--d-text-muted)" }}>
                Автор: {inspectBooking.userEmail}
              </div>
            )}
            {inspectBooking.canManage && (
              <div className="flex gap-2 flex-wrap">
                <button
                  className="btn btn-primary text-sm"
                  onClick={() => { setEditBooking(inspectBooking); setInspectBooking(null); setFormError(""); }}
                >
                  Редактировать
                </button>
                <button
                  className="btn text-sm"
                  style={{ color: "var(--d-error)" }}
                  onClick={() => requestDelete(inspectBooking.id)}
                >
                  Удалить
                </button>
              </div>
            )}
          </div>
        </Modal>
      )}

      {/* Delete confirmation modal */}
      <Modal open={!!confirmDeleteId} onClose={() => setConfirmDeleteId(null)}>
        {confirmDeleteId && (() => {
          const bk = bookings.find(b => b.id === confirmDeleteId);
          return (
            <div className="space-y-5">
              <div>
                <h2 className="text-lg font-bold mb-1" style={{ color: "var(--d-text)" }}>
                  Удалить бронирование?
                </h2>
                {bk && (
                  <p className="text-sm" style={{ color: "var(--d-text-muted)" }}>
                    {new Date(bk.start).toLocaleString("ru", { day: "2-digit", month: "long", hour: "2-digit", minute: "2-digit" })}
                    {" – "}
                    {new Date(bk.end).toLocaleTimeString("ru", { hour: "2-digit", minute: "2-digit" })}
                    {" · "}{bk.title}
                  </p>
                )}
              </div>
              <div className="flex gap-3">
                <button
                  className="btn flex-1"
                  onClick={() => setConfirmDeleteId(null)}
                >
                  Отмена
                </button>
                <button
                  className="btn flex-1 font-semibold"
                  style={{ background: "var(--d-error)", color: "#fff", borderColor: "var(--d-error)" }}
                  onClick={() => handleDelete(confirmDeleteId)}
                >
                  Да, удалить
                </button>
              </div>
            </div>
          );
        })()}
      </Modal>

      {/* Edit booking modal */}
      <Modal open={!!editBooking} onClose={() => { setEditBooking(null); setFormError(""); }}>
        {editBooking && (
          <BookingWizard
            activeRooms={activeRooms}
            onSubmit={handleUpdate}
            onCancel={() => { setEditBooking(null); setFormError(""); }}
            formLoading={formLoading}
            formError={formError}
            initialBooking={editBooking}
          />
        )}
      </Modal>
    </div>
  );
}
