import React, { useEffect, useRef, useState } from "react";
import { Trash2, Plus, XCircle, CheckCircle, ShieldCheck } from "lucide-react";
import type { AdminEntry } from "../../types/bookings";
import { fetchAdmins, addAdmin, removeAdmin } from "../../api/admins";
import { normalizeApiError } from "../../utils/errors";

interface Props {
  isSuperAdmin: boolean;
}

export function AdminsTab({ isSuperAdmin }: Props) {
  const [admins, setAdmins] = useState<AdminEntry[]>([]);
  const [email, setEmail] = useState("");
  const [addError, setAddError] = useState("");
  const [addSuccess, setAddSuccess] = useState(false);
  const [addLoading, setAddLoading] = useState(false);
  const [confirmEmail, setConfirmEmail] = useState<string | null>(null);
  const [removeError, setRemoveError] = useState("");
  const [removing, setRemoving] = useState(false);
  const successTimer = useRef<number | null>(null);

  async function load() {
    try {
      setAdmins(await fetchAdmins());
    } catch (e: any) {
      console.error("[AdminsTab] load:", e);
    }
  }

  useEffect(() => { load(); }, []);
  useEffect(() => () => { if (successTimer.current) clearTimeout(successTimer.current); }, []);

  async function handleAdd() {
    setAddError("");
    setAddSuccess(false);
    if (successTimer.current) clearTimeout(successTimer.current);

    const e = email.trim().toLowerCase();
    if (!/@(edu\.hse\.ru|hse\.ru)$/.test(e)) {
      setAddError("Принимаются только адреса @hse.ru и @edu.hse.ru");
      return;
    }
    setAddLoading(true);
    try {
      await addAdmin(e);
      setEmail("");
      setAddSuccess(true);
      successTimer.current = window.setTimeout(() => setAddSuccess(false), 3000);
      await load();
    } catch (err: any) {
      setAddError(normalizeApiError(err));
    } finally {
      setAddLoading(false);
    }
  }

  async function handleRemoveConfirmed() {
    if (!confirmEmail) return;
    setRemoving(true);
    setRemoveError("");
    try {
      await removeAdmin(confirmEmail);
      setConfirmEmail(null);
      await load();
    } catch (ex: any) {
      setRemoveError(normalizeApiError(ex));
      setConfirmEmail(null);
    } finally {
      setRemoving(false);
    }
  }

  if (!isSuperAdmin) {
    return (
      <div className="flex items-center gap-2 py-6 text-sm" style={{ color: "var(--d-text-sec)" }}>
        <ShieldCheck className="h-4 w-4" style={{ color: "var(--d-primary)" }} />
        Управление администраторами доступно только суперадмину.
      </div>
    );
  }

  return (
    <div className="space-y-5 pt-5">
      {/* Add form */}
      <div
        className="rounded-2xl border p-4 space-y-3"
        style={{ background: "var(--d-panel)", borderColor: "var(--d-border)" }}
      >
        <div className="font-semibold text-sm" style={{ color: "var(--d-text)" }}>Добавить администратора</div>

        {addError && (
          <div className="banner-error">
            <XCircle className="h-4 w-4 shrink-0 mt-0.5" />
            {addError}
          </div>
        )}

        {addSuccess && (
          <div className="banner-ok flex items-center gap-2 rounded-xl px-3 py-2 text-sm"
            style={{ background: "rgba(134,156,39,0.12)", color: "var(--d-primary)", border: "1px solid rgba(134,156,39,0.25)" }}>
            <CheckCircle className="h-4 w-4 shrink-0" />
            Администратор добавлен
          </div>
        )}

        <label className="flex flex-col gap-1 text-sm">
          <span className="lbl">Email (@hse.ru / @edu.hse.ru)</span>
          <input
            className="field"
            type="email"
            placeholder="ivanov@edu.hse.ru"
            value={email}
            onChange={(e) => { setEmail(e.target.value); setAddError(""); }}
            onKeyDown={(e) => e.key === "Enter" && handleAdd()}
          />
        </label>

        <button className="btn btn-primary text-sm" onClick={handleAdd} disabled={addLoading}>
          <Plus className="h-4 w-4" />
          {addLoading ? "Добавление…" : "Добавить"}
        </button>
      </div>

      {/* Remove error (floats above list) */}
      {removeError && (
        <div className="banner-error flex items-center gap-2">
          <XCircle className="h-4 w-4 shrink-0" />
          <span className="flex-1">{removeError}</span>
          <button
            className="shrink-0 opacity-60 hover:opacity-100 transition-opacity"
            onClick={() => setRemoveError("")}
            aria-label="Закрыть"
          >
            ×
          </button>
        </div>
      )}

      {/* Admin list */}
      <div className="space-y-2">
        {admins.length === 0 ? (
          <div className="text-sm py-4 text-center" style={{ color: "var(--d-text-muted)" }}>
            Нет администраторов
          </div>
        ) : admins.map((a) => (
          <div
            key={a.email}
            className="flex items-center justify-between gap-3 rounded-xl border px-3 py-2.5 transition-all duration-200"
            style={{ background: "var(--d-surface)", borderColor: "var(--d-border)" }}
          >
            <div className="min-w-0">
              <div className="font-medium text-sm truncate" style={{ color: "var(--d-text)" }}>
                {a.email}
              </div>
            </div>

            {(
              confirmEmail === a.email ? (
                <div className="flex items-center gap-1.5 shrink-0">
                  <span className="text-xs whitespace-nowrap" style={{ color: "var(--d-text-sec)" }}>Удалить?</span>
                  <button
                    className="btn text-xs px-2.5 py-1"
                    style={{ color: "var(--d-error)", borderColor: "var(--d-error)" }}
                    onClick={handleRemoveConfirmed}
                    disabled={removing}
                  >
                    {removing ? "…" : "Да"}
                  </button>
                  <button
                    className="btn text-xs px-2.5 py-1"
                    onClick={() => setConfirmEmail(null)}
                    disabled={removing}
                  >
                    Нет
                  </button>
                </div>
              ) : (
                <button
                  className="icon-btn shrink-0"
                  style={{ color: "var(--d-error)" }}
                  onClick={() => { setRemoveError(""); setConfirmEmail(a.email); }}
                  title="Удалить"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              )
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
