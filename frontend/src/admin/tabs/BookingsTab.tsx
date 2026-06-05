import React, { useEffect, useState } from "react";
import { Trash2, Zap, XCircle } from "lucide-react";
import type { Booking, RoomInfo } from "../../types/bookings";
import { fetchBookings, deleteBooking, forceCreateBooking } from "../../api/bookings";
import { normalizeApiError } from "../../utils/errors";
import { DatePicker } from "../../components/pickers/DatePicker";
import { TimePicker } from "../../components/time/TimePicker";
import { toYMD, pad2 } from "../../utils/date";
import { Modal } from "../../components/ui/Modal";

function fmtTime(d: string) {
  return new Date(d).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

function fmtDateHeader(ymd: string) {
  const d = new Date(ymd + "T00:00:00");
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "long", weekday: "long" });
}

interface Props {
  rooms: RoomInfo[];
  onSuccess?: () => void;
}

export function BookingsTab({ rooms, onSuccess }: Props) {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const now = new Date();
  const [fp, setFp] = useState({
    room: rooms[0]?.number ?? 21,
    date: toYMD(now),
    startH: now.getHours() + 1,
    startM: 0,
    endH: now.getHours() + 2,
    endM: 0,
    title: "",
    description: "",
    isPrivate: false,
  });
  const [fpError, setFpError] = useState("");
  const [fpLoading, setFpLoading] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    setError("");
    try {
      setBookings(await fetchBookings());
    } catch (e: any) {
      setError("Ошибка загрузки. Попробуйте позже.");
      console.error("[BookingsTab] load:", e);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function handleDelete(id: string) {
    setConfirmDeleteId(null);
    try {
      await deleteBooking(id);
      setBookings((b) => b.filter((x) => x.id !== id));
    } catch (e: any) {
      setError(normalizeApiError(e));
      console.error("[BookingsTab] delete:", e);
    }
  }

  async function handleForcePush() {
    setFpError("");
    if (!fp.title) {
      setFpError("Заполните название");
      return;
    }
    const buildISO = (h: number, m: number) => {
      const d = new Date(fp.date);
      d.setHours(h, m, 0, 0);
      return d.toISOString();
    };
    const startISO = buildISO(fp.startH, fp.startM);
    const endISO   = buildISO(fp.endH,   fp.endM);
    if (new Date(endISO) <= new Date(startISO)) {
      setFpError("Конец должен быть позже начала");
      return;
    }
    setFpLoading(true);
    try {
      await forceCreateBooking({
        room: fp.room,
        start: startISO,
        end: endISO,
        title: fp.title,
        description: fp.description || undefined,
        isPrivate: fp.isPrivate,
      });
      setFp(f => ({ ...f, title: "", description: "" }));
      await load();
      onSuccess?.();
    } catch (e: any) {
      setFpError(normalizeApiError(e));
    } finally {
      setFpLoading(false);
    }
  }

  // Group bookings by date, ascending
  const sorted = [...bookings].sort((a, b) => +new Date(a.start) - +new Date(b.start));
  const groups = new Map<string, typeof sorted>();
  for (const b of sorted) {
    const key = toYMD(new Date(b.start));
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key)!.push(b);
  }
  const groupKeys = [...groups.keys()];

  return (
    <>
    <div className="space-y-5 pt-5">
      {/* Force-push */}
      <div
        className="rounded-2xl border p-4 space-y-3"
        style={{ background: "rgba(201,148,58,0.06)", borderColor: "rgba(201,148,58,0.3)" }}
      >
        <div className="flex items-center gap-2 font-semibold text-sm" style={{ color: "var(--d-warning)" }}>
          <Zap className="h-4 w-4" />
          Форс-пуш брони
          <span className="text-xs font-normal opacity-70 ml-1">(обходит все правила, кроме перекрытия)</span>
        </div>

        {fpError && (
          <div className="banner-error">
            <XCircle className="h-4 w-4 shrink-0 mt-0.5" />
            {fpError}
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <label className="flex flex-col gap-1 text-sm">
            <span className="lbl">Комната</span>
            <select className="field" value={fp.room}
              onChange={(e) => setFp({ ...fp, room: Number(e.target.value) })}>
              {rooms.filter(r => r.isActive).map(r => (
                <option key={r.number} value={r.number}>{r.name}</option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-sm">
            <span className="lbl">Название *</span>
            <input className="field" value={fp.title}
              onChange={(e) => setFp({ ...fp, title: e.target.value })} />
          </label>

          {/* Single date + start/end times */}
          <div className="col-span-2 space-y-2">
            <span className="lbl">Дата *</span>
            <DatePicker value={fp.date} onChange={(v) => setFp(f => ({ ...f, date: v }))} />
          </div>

          <div className="space-y-1">
            <span className="lbl text-sm">Начало *</span>
            <TimePicker
              label="Время начала"
              dateStr={fp.date}
              compact
              value={`${pad2(fp.startH)}:${pad2(fp.startM)}`}
              onChange={(hm) => {
                const [h, m] = hm.split(":").map(Number);
                setFp(f => ({ ...f, startH: h, startM: m }));
              }}
            />
          </div>
          <div className="space-y-1">
            <span className="lbl text-sm">Конец *</span>
            <TimePicker
              label="Время конца"
              dateStr={fp.date}
              compact
              value={`${pad2(fp.endH)}:${pad2(fp.endM)}`}
              onChange={(hm) => {
                const [h, m] = hm.split(":").map(Number);
                setFp(f => ({ ...f, endH: h, endM: m }));
              }}
            />
          </div>

          <label className="flex flex-col gap-1 text-sm col-span-2">
            <span className="lbl">Описание</span>
            <input className="field" value={fp.description}
              onChange={(e) => setFp({ ...fp, description: e.target.value })} />
          </label>

          <div className="col-span-2">
            <span className="lbl block mb-1.5">Тип</span>
            <div className="inline-flex rounded-xl p-1 gap-1" style={{ background: "var(--d-surface)" }}>
              {[{ v: false, l: "Публичное" }, { v: true, l: "Частное" }].map(({ v, l }) => (
                <button
                  key={String(v)}
                  type="button"
                  onClick={() => setFp(f => ({ ...f, isPrivate: v }))}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-200"
                  style={fp.isPrivate === v ? {
                    background: "var(--d-primary)", color: "white",
                  } : {
                    background: "transparent", color: "var(--d-text-muted)",
                  }}
                >
                  {l}
                </button>
              ))}
            </div>
          </div>
        </div>

        <button className="btn btn-primary text-sm" onClick={handleForcePush} disabled={fpLoading}>
          <Zap className="h-4 w-4" />
          {fpLoading ? "Создание…" : "Форс-пуш"}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="banner-error">
          <XCircle className="h-4 w-4 shrink-0 mt-0.5" />
          {error}
        </div>
      )}

      {/* Bookings list grouped by date */}
      <div className="space-y-1">
        <div className="text-xs font-medium uppercase tracking-wide mb-3" style={{ color: "var(--d-text-sec)" }}>
          {sorted.length} броней
        </div>
        {loading ? (
          <div className="text-sm" style={{ color: "var(--d-text-muted)" }}>Загрузка…</div>
        ) : sorted.length === 0 ? (
          <div className="text-sm py-4 text-center" style={{ color: "var(--d-text-muted)" }}>Бронирований нет</div>
        ) : groupKeys.map((key) => (
          <div key={key} className="space-y-1.5 mb-4">
            <div
              className="text-xs font-semibold capitalize py-2 sticky top-0 z-10 -mx-5 px-5"
              style={{ color: "var(--d-text-sec)", background: "var(--d-surface)" }}
            >
              {fmtDateHeader(key)}
            </div>
            {groups.get(key)!.map((b) => (
              <div
                key={b.id}
                className="flex items-center justify-between gap-3 rounded-xl border px-3 py-2.5 transition-all duration-200 hover:shadow-sm"
                style={{ background: "var(--d-surface)", borderColor: "var(--d-border)" }}
              >
                <div className="min-w-0 flex-1">
                  <div className="font-medium text-sm truncate" style={{ color: "var(--d-text)" }}>{b.title}</div>
                  <div className="text-xs mt-0.5" style={{ color: "var(--d-text-sec)" }}>
                    Комн. {b.room} · {fmtTime(b.start)} – {fmtTime(b.end)}
                    {b.isPrivate && " · ЧП"}
                  </div>
                  <div className="text-xs truncate" style={{ color: "var(--d-text-muted)" }}>
                    {b.userEmail || (b.telegramId ? `@${b.telegramId}` : "—")}
                  </div>
                </div>
                <button
                  className="icon-btn shrink-0"
                  style={{ color: "var(--d-error)" }}
                  onClick={() => setConfirmDeleteId(b.id)}
                  title="Удалить"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>

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
                  {fmtTime(bk.start)} – {fmtTime(bk.end)} · {bk.title}
                </p>
              )}
            </div>
            <div className="flex gap-3">
              <button className="btn flex-1" onClick={() => setConfirmDeleteId(null)}>
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
    </>
  );
}
