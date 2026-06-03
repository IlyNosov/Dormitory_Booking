import React, {useEffect, useMemo, useState} from "react";
import {AlertTriangle, Filter, Lock, Menu, Unlock} from "lucide-react";

import type {Bookings, CreateBookingPayload, RoomFilter, ViewMode, VisFilter} from "../types/bookings";
import {cn} from "../utils/cn";
import {fmtRange, fromYMD, msToHuman, pad2, toYMD} from "../utils/date";
import {isFriOrSat, isQuiet} from "../utils/rules";
import {ROOM_COLORS} from "../utils/rooms";

import {AppHeader} from "./AppHeader";
import {DateFilter} from "../components/filters/DateFilter";
import {TimePicker} from "../components/time/TimePicker";
import {CardPattern} from "../components/patterns/CardPattern";
import {HoverHint} from "../components/ui/HoverHint";
import {Modal} from "../components/ui/Modal";
import {Popover} from "../components/ui/Popover";
import {DatePicker} from "../components/pickers/DatePicker";
import {useAdmin} from "../admin/useAdmin";
import {AdminModal} from "../admin/AdminModal";
import * as api from "../api/bookings";
import {setAdminToken} from "../api/bookings";
import { RulesModal } from "../components/ui/RulesModal";
import {BookingsTableView} from "../table/BookingsTableView";

function SelectInner<T extends string | number>({
                                                    label,
                                                    value,
                                                    options,
                                                    onChange,
                                                    format = String,
                                                }: {
    label?: string;
    value: T;
    options: readonly T[];
    onChange: (v: T) => void;
    format?: (v: T) => string;
}) {
    const [open, setOpen] = useState(false);
    const ref = React.useRef<HTMLButtonElement | null>(null);
    return (
        <label className="flex flex-col gap-1 text-sm">
            {label && <span className="lbl">{label}</span>}
            <button ref={ref} type="button" className="field flex items-center justify-between"
                    onClick={() => setOpen((o) => !o)}>
                <span>{format(value)}</span>
                <span className="opacity-60">▾</span>
            </button>
            <Popover open={open} onClose={() => setOpen(false)} anchor={ref.current}>
                <div className="max-h-64 overflow-auto rounded-xl">
                    {options.map((opt) => (
                        <button
                            key={String(opt)}
                            type="button"
                            onClick={() => {
                                onChange(opt);
                                setOpen(false);
                            }}
                            className={cn(
                                "block w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800",
                                String(opt) === String(value) && "bg-blue-50 dark:bg-blue-950/30",
                            )}
                        >
                            {format(opt)}
                        </button>
                    ))}
                </div>
            </Popover>
        </label>
    );
}

export function App() {

    const admin = useAdmin();
    const [adminOpen, setAdminOpen] = useState(false);

    useEffect(() => {
        setAdminToken(admin.token);
    }, [admin.token]);

    const [view, setView] = useState<ViewMode>("cards");
    const [viewKey, setViewKey] = useState(0);
    const [rulesOpen, setRulesOpen] = useState(false);
    const [bookings, setBookings] = useState<Bookings[] | null>(null);
    const [loading, setLoading] = useState(false);
    const [adding, setAdding] = useState(false);
    const [errMsg, setErrMsg] = useState("");
    const [nowTs, setNowTs] = useState(Date.now());
    const [inspect, setInspect] = useState<Bookings | null>(null);

    useEffect(() => {
        const t = setInterval(() => setNowTs(Date.now()), 30000);
        return () => clearInterval(t);
    }, []);

    const startOfToday = new Date();
    startOfToday.setHours(0, 0, 0, 0);
    const [fromDate, setFromDate] = useState(toYMD(startOfToday));
    const [days, setDays] = useState(7);
    const [roomFilter, setRoomFilter] = useState<RoomFilter>("all");
    const [visFilter, setVisFilter] = useState<VisFilter>("all");

    const dS = new Date(Date.now() + 30 * 60 * 1000);
    const dE = new Date(Date.now() + 90 * 60 * 1000);
    const step5 = (n: number) => pad2((Math.round(n / 5) * 5) % 60);

    const [form, setForm] = useState({
        date: toYMD(dS),
        startTime: `${pad2(dS.getHours())}:${step5(dS.getMinutes())}`,
        endTime: `${pad2(dE.getHours())}:${step5(dE.getMinutes())}`,
        room: 21 as 21 | 132 | 256,
        title: "",
        telegramId: "",
        isPrivate: false,
        description: "",
    });

    async function fetchData() {
        setLoading(true);
        try {
            const data = await api.fetchBookings();
            setBookings(data);
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        fetchData();
    }, []);

    const sorted = useMemo(
        () => (bookings ? [...bookings].sort((a, b) => +new Date(a.start) - +new Date(b.start)) : []),
        [bookings],
    );

    const {futureByDay, pastFiltered, futureCount} = useMemo(() => {
        const sw = +new Date(`${fromDate}T00:00`);
        const ew = sw + days * 24 * 3600_000;

        const fits = (b: Bookings) => {
            if (roomFilter !== "all" && b.room !== roomFilter) return false;
            if (visFilter === "public" && b.isPrivate) return false;
            if (visFilter === "private" && !b.isPrivate) return false;
            return true;
        };

        const future = new Map<string, Bookings[]>();
        const past: Bookings[] = [];
        let fcount = 0;

        for (const b of sorted) {
            const s = +new Date(b.start);
            const e = +new Date(b.end);

            if (e < nowTs) {
                if (fits(b)) past.push(b);
                continue;
            }

            const overlap = !(e <= sw || s >= ew);
            if (!overlap || !fits(b)) continue;

            const key = toYMD(new Date(b.start));
            (future.get(key) ?? future.set(key, []).get(key)!).push(b);
            fcount++;
        }

        for (const [, v] of future) v.sort((a, b) => +new Date(a.start) - +new Date(b.start));
        past.sort((a, b) => +new Date(b.end) - +new Date(a.end));

        return {futureByDay: future, pastFiltered: past, futureCount: fcount};
    }, [sorted, fromDate, days, roomFilter, visFilter, nowTs]);

    function computeStartEnd() {
        const start = new Date(`${form.date}T${form.startTime}:00`);
        const endSame = new Date(`${form.date}T${form.endTime}:00`);

        const [sh, sm] = form.startTime.split(":").map(Number);
        const [eh, em] = form.endTime.split(":").map(Number);

        let end = endSame;
        const startMin = sh * 60 + sm;
        const endMin = eh * 60 + em;

        if (endMin < startMin) {
            const base = fromYMD(form.date);
            if (!(isFriOrSat(base) && endMin <= 60)) {
                throw new Error("Перенос через полночь допустим только Пт/Сб до 01:00.");
            }
            end = new Date(+endSame + 24 * 3600_000);
        }
        return {start, end};
    }

    function validate(): string | null {
        if (!form.title.trim()) return "Заполни поле «Название».";
        if (!form.telegramId.trim()) return "Укажи Telegram ID.";

        let start: Date, end: Date;
        try {
            ({start, end} = computeStartEnd());
        } catch (e: any) {
            return String(e?.message || e);
        }

        if (isNaN(+start) || isNaN(+end)) return "Укажи корректные дату и время.";
        if (end <= start) return "Время окончания должно быть позже начала.";

        const dur = +end - +start;
        if (form.isPrivate && dur > 3 * 3600_000) return "Частная бронь не может длиться дольше 3 часов.";

        if (form.isPrivate) {
            for (let t = new Date(start); t < end; t = new Date(+t + 15 * 60_000)) {
                if (isQuiet(t)) return "Частная бронь недоступна с 23:00 до 06:00 (до 01:00 ночи Пт-Сб и Сб-Вс).";
            }
        }
        return null;
    }

    async function handleCreate() {
        setErrMsg("");
        const e = validate();
        if (e) return setErrMsg(e);

        const {start, end} = computeStartEnd();
        const payload: CreateBookingPayload = {
            start: start.toISOString(),
            end: end.toISOString(),
            room: form.room,
            title: form.title.trim(),
            telegramId: form.telegramId.trim(),
            isPrivate: form.isPrivate,
            description: form.description?.trim() || undefined,
        };

        try {
            const created = await api.createBooking(payload);
            setBookings((p) => (p ? [created, ...p] : [created]));
            setForm((f) => ({...f, title: "", description: ""}));
            setAdding(false);
        } catch (err: any) {
            setErrMsg(String(err?.message || err));
        }
    }

    async function handleDelete(id: string, owner?: string) {
        try {
            await api.deleteBooking(id, owner);
            setBookings((p) => (p ? p.filter((b) => b.id !== id) : p));
        } catch (e: any) {
            setErrMsg(String(e?.message || e));
            await fetchData();
        }
    }

    return (
        <div
            className="min-h-screen bg-zinc-50 text-zinc-900 dark:bg-[color:var(--d-bg)] dark:text-zinc-100 transition-colors">
            <AppHeader
                loading={loading}
                view={view}
                isAdmin={admin.isAdmin}
                onToggleAdd={() => setAdding((v) => !v)}
                onRefresh={fetchData}
                onToggleView={() => {
                    setView((v) => (v === "cards" ? "table" : "cards"));
                    setViewKey((k) => k + 1);
                }}
                onAdminClick={() => setAdminOpen(true)}
                onAdminLogout={admin.logout}
                onRulesClick={() => setRulesOpen(true)}
            />


            <main className="mx-auto max-w-6xl px-4 py-6">
                <div
                    className={cn(
                        "mb-6 overflow-hidden transition-[max-height,opacity,transform] duration-300",
                        adding ? "max-h-[1400px] opacity-100 translate-y-0" : "max-h-0 opacity-0 -translate-y-2",
                    )}
                >
                    <div
                        className="rounded-2xl border border-zinc-200 dark:border-[color:var(--d-border)] bg-white/80 dark:bg-[color:var(--d-card)] backdrop-blur p-4">
                        {errMsg && (
                            <div
                                className="mb-3 flex items-start gap-2 rounded-xl border border-red-300/50 bg-red-50 dark:bg-rose-950/20 text-red-700 dark:text-rose-300 px-3 py-2 text-sm">
                                <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5"/>
                                <div>{errMsg}</div>
                            </div>
                        )}

                        <div className="grid gap-6 md:grid-cols-[320px_1fr]">
                            <DatePicker
                                value={form.date}
                                onChange={(v) => setForm((f) => ({...f, date: v}))}
                            />

                            <div className="grid sm:grid-cols-2 gap-6 items-center">
                                <TimePicker
                                    label="Время начала"
                                    dateStr={form.date}
                                    value={form.startTime}
                                    onChange={(hm) => setForm((f) => ({...f, startTime: hm}))}
                                />

                                <TimePicker
                                    label="Время окончания"
                                    dateStr={form.date}
                                    value={form.endTime}
                                    onChange={(hm) => setForm((f) => ({...f, endTime: hm}))}
                                />
                            </div>
                        </div>

                        <div className="mt-6 grid md:grid-cols-2 gap-4">
                            <SelectInner label="Комната" value={form.room} options={[21, 256, 132] as const}
                                         onChange={(v) => setForm((f) => ({...f, room: v}))}
                                         format={(v) => `Комната ${v}`}/>

                            <label className="flex flex-col gap-1 text-sm">
                                <span className="lbl">Название</span>
                                <input className="field" placeholder="Название мероприятия" value={form.title}
                                       onChange={(e) => setForm((f) => ({...f, title: e.target.value}))}/>
                            </label>

                            <label className="flex items-center gap-3 text-sm select-none">
                                <button type="button" onClick={() => setForm((f) => ({...f, isPrivate: !f.isPrivate}))}
                                        className={cn("switch", form.isPrivate && "switch-on")}
                                        aria-pressed={form.isPrivate}/>
                                <span
                                    className={cn("font-medium tracking-wide", form.isPrivate ? "text-rose-600 dark:text-rose-300" : "text-emerald-700 dark:text-emerald-300")}>
                  {form.isPrivate ? "Частная бронь" : "Публичная бронь"}
                </span>
                            </label>

                            <label className="flex flex-col gap-1 text-sm">
                                <span className="lbl">Telegram ID</span>
                                <input className="field" placeholder="@username" value={form.telegramId}
                                       onChange={(e) => setForm((f) => ({...f, telegramId: e.target.value}))}/>
                            </label>

                            <label className="md:col-span-2 flex flex-col gap-1 text-sm">
                                <span className="lbl">Описание (необязательно)</span>
                                <textarea className="field min-h-[74px]" placeholder="Подробности брони…"
                                          value={form.description}
                                          onChange={(e) => setForm((f) => ({...f, description: e.target.value}))}/>
                            </label>
                        </div>

                        <div className="mt-4 flex gap-2">
                            <button onClick={handleCreate} className="btn btn-primary">
                                Создать
                            </button>
                            <button onClick={() => setAdding(false)} className="btn">
                                Отмена
                            </button>
                        </div>
                    </div>
                </div>

                {/* фильтры */}
                {view === "cards" ? (
                    <section className="mb-5 fade-wrap">
                        <div
                            className="rounded-2xl border border-zinc-200 dark:border-[color:var(--d-border)] bg-white/75 dark:bg-[color:var(--d-card)] backdrop-blur p-4">
                            <div className="grid gap-4 md:grid-cols-[320px_1fr]">
                                <div className="grid grid-rows-[auto_auto] gap-2">
                                    <span className="text-xs uppercase tracking-wide text-zinc-600 dark:text-zinc-400">
                                        фильтры
                                    </span>
                                    <div className="flex items-end gap-6">
                                        <DateFilter value={fromDate} onChange={setFromDate}/>
                                        <div className="w-16">
                                            <SelectInner
                                                label="Дней"
                                                value={days as any}
                                                options={[1, 3, 5, 7, 10, 14] as const}
                                                onChange={(v) => setDays(Number(v))}
                                            />
                                        </div>
                                    </div>
                                </div>

                                <div className="grid grid-rows-[auto_auto_auto] gap-3 justify-items-end">
                                    <div className="flex flex-wrap justify-end gap-2">
                                        <button onClick={() => setRoomFilter("all")}
                                                className={cn("btn", roomFilter === "all" && "btn-active")}>
                                            Все комнаты
                                        </button>
                                        {[21, 256, 132].map((v) => (
                                            <button key={`rm-${v}`} onClick={() => setRoomFilter(v as any)}
                                                    className={cn("btn", roomFilter === v && "btn-active")}>
                                                {`Комната ${v}`}
                                            </button>
                                        ))}
                                    </div>

                                    <div className="flex flex-wrap justify-end gap-2">
                                        {(["all", "public", "private"] as const).map((v) => (
                                            <button key={`vis-${v}`} onClick={() => setVisFilter(v)}
                                                    className={cn("btn", visFilter === v && "btn-active")}>
                                                {v === "all" ? "Все" : v === "public" ? "Публичные" : "Частные"}
                                            </button>
                                        ))}
                                    </div>
                                    <div/>
                                </div>
                            </div>

                        </div>
                    </section>
                ) : null}

                <div key={viewKey} className="fade-wrap">
                    {view === "cards" ? (
                        <>
                            <section className="space-y-6">
                                {[...futureByDay.keys()].sort().map((dayKey) => {
                                    const items = futureByDay.get(dayKey)!;
                                    const ruDate = new Date(dayKey).toLocaleDateString("ru-RU", {
                                        weekday: "long",
                                        day: "2-digit",
                                        month: "long"
                                    });

                                    return (
                                        <div key={dayKey}>
                                            <h2 className="mb-3 text-lg font-semibold capitalize">{ruDate}</h2>
                                            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                                                {items.map((b) => {
                                                    const s = +new Date(b.start),
                                                        e = +new Date(b.end);
                                                    const active = nowTs >= s && nowTs <= e;
                                                    const total = Math.max(1, e - s);
                                                    const progress = Math.min(100, Math.max(0, ((nowTs - s) / total) * 100));
                                                    const rc = ROOM_COLORS[b.room];

                                                    const card = (
                                                        <article
                                                            className={cn("card group relative overflow-hidden transition-all duration-300", active ? `ring-2 ${rc.ring} ${rc.bg}` : "opacity-90 hover:opacity-100")}
                                                            onClick={() => b.description && setInspect(b)}
                                                        >
                                                            <CardPattern room={b.room} active={active}/>
                                                            {active && <div
                                                                className={cn("absolute inset-0 z-[1] pointer-events-none", rc.bg)}/>}
                                                            <div
                                                                className="flex items-start justify-between mb-2 gap-3">
                                                                <span
                                                                    className="text-xs font-medium tracking-wide text-zinc-500 dark:text-zinc-400 mt-1">{fmtRange(b.start, b.end)}</span>
                                                                <div className="flex items-center gap-2">
                                  <span
                                      className={cn("inline-flex items-center gap-1 rounded-full px-2 py-1 text-[11px] font-medium", b.isPrivate ? "badge-priv" : "badge-pub")}>
                                    {b.isPrivate ? <Lock className="h-3.5 w-3.5"/> : <Unlock className="h-3.5 w-3.5"/>}
                                      {b.isPrivate ? "Частная" : "Публичная"}
                                  </span>
                                                                    {b.description &&
                                                                        <Menu className="h-4 w-4 opacity-80"/>}
                                                                    {admin.isAdmin && (
                                                                        <button
                                                                            onClick={(e) => {
                                                                                e.stopPropagation();
                                                                                handleDelete(b.id, b.telegramId);
                                                                            }}
                                                                            className="icon-btn"
                                                                            title="Удалить бронь"
                                                                            aria-label="Удалить бронь"
                                                                        >
                                                                            <svg xmlns="http://www.w3.org/2000/svg"
                                                                                 viewBox="0 0 24 24" width="16"
                                                                                 height="16" fill="none"
                                                                                 stroke="currentColor" strokeWidth="2"
                                                                                 strokeLinecap="round"
                                                                                 strokeLinejoin="round">
                                                                                <polyline points="3 6 5 6 21 6"/>
                                                                                <path
                                                                                    d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                                                                                <path d="M10 11v6"/>
                                                                                <path d="M14 11v6"/>
                                                                                <path
                                                                                    d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
                                                                            </svg>
                                                                        </button>
                                                                    )}
                                                                </div>
                                                            </div>

                                                            <h3 className="text-base font-semibold leading-tight mb-1 line-clamp-2">{b.title}</h3>

                                                            <div
                                                                className="mt-2 grid grid-cols-2 gap-2 text-sm text-zinc-600 dark:text-zinc-300">
                                                                <div
                                                                    className="rounded-xl bg-zinc-100 dark:bg-[color:var(--d-panel)] p-2">
                                                                    <div
                                                                        className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-400">Комната
                                                                    </div>
                                                                    <div className="font-medium">{b.room}</div>
                                                                </div>
                                                                <div
                                                                    className="rounded-xl bg-zinc-100 dark:bg-[color:var(--d-panel)] p-2">
                                                                    <div
                                                                        className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-400">Telegram
                                                                    </div>
                                                                    <div
                                                                        className="font-medium break-all">{b.telegramId}</div>
                                                                </div>
                                                            </div>

                                                            <div className="mt-4 mb-1">
                                                                <div
                                                                    className="h-2 rounded-full bg-zinc-200 dark:bg-[color:var(--d-panel)] w-full overflow-hidden">
                                                                    {active && (
                                                                        <div
                                                                            className="h-full bg-blue-500 transition-[width] duration-500"
                                                                            style={{width: `${progress}%`}}
                                                                            title={`Прогресс: ${msToHuman(nowTs - s)} из ${msToHuman(total)}`}
                                                                        />
                                                                    )}
                                                                </div>
                                                            </div>
                                                        </article>
                                                    );

                                                    return b.description ? (
                                                        <HoverHint key={b.id} hint="Нажмите, чтобы увидеть описание">
                                                            {card}
                                                        </HoverHint>
                                                    ) : (
                                                        <div key={b.id}>{card}</div>
                                                    );
                                                })}
                                            </div>
                                        </div>
                                    );
                                })}
                            </section>

                            {futureCount === 0 && (
                                <div className="mt-4 text-sm text-zinc-600 dark:text-zinc-400">
                                    {roomFilter === "all" && visFilter === "all" ? "На заданный период брони не обнаружены." : "По заданному фильтру брони не найдены."}
                                </div>
                            )}
                        </>
                    ) : (
                        <BookingsTableView
                            dayKey={fromDate}
                            bookings={sorted}
                            onInspect={(b: Bookings) => setInspect(b)}
                        />
                    )}
                </div>
            </main>

            <Modal open={!!inspect} onClose={() => setInspect(null)}>
                {inspect && (
                    <div className="space-y-3">
                        <div className="text-lg font-semibold">{inspect.title}</div>
                        <div
                            className="text-sm text-zinc-600 dark:text-zinc-300">{fmtRange(inspect.start, inspect.end)}</div>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                            <div className="rounded-xl bg-zinc-100 p-2 dark:bg-[color:var(--d-panel)]">
                                <div
                                    className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-400">Комната
                                </div>
                                <div className="font-medium">{inspect.room}</div>
                            </div>
                            <div className="rounded-xl bg-zinc-100 p-2 dark:bg-[color:var(--d-panel)]">
                                <div
                                    className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-400">Telegram
                                </div>
                                <div className="font-medium break-all">{inspect.telegramId}</div>
                            </div>
                        </div>
                        {inspect.description && (
                            <div
                                className="rounded-xl border border-zinc-200/70 p-3 text-[14px] leading-5 dark:border-[color:var(--d-border)]">{inspect.description}</div>
                        )}
                    </div>
                )}
            </Modal>
            <AdminModal open={adminOpen} onClose={() => setAdminOpen(false)} onLogin={(t) => admin.login(t)}/>
            <RulesModal open={rulesOpen} onClose={() => setRulesOpen(false)}/>
        </div>
    );
}
