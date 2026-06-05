import React, { useRef, useState } from "react";
import { Shield, LogOut, LogIn, Bug, Menu, X, DoorOpen, BookMarked } from "lucide-react";
import type { User, ViewMode } from "../types/bookings";
import { cn } from "../utils/cn";
import { Popover } from "../components/ui/Popover";
import { sendBugReport } from "../api/bugs";

interface Props {
  view: ViewMode;
  user: User | null | undefined;
  myBookings: boolean;
  onToggleAdd: () => void;
  onToggleView: (v: ViewMode) => void;
  onToggleMyBookings: () => void;
  onLoginClick: () => void;
  onAdminClick: () => void;
  onLogout: () => void;
  onRulesClick: () => void;
  onRoomsClick: () => void;
}

export function AppHeader({
  view, user, myBookings,
  onToggleAdd, onToggleView, onToggleMyBookings,
  onLoginClick, onAdminClick, onLogout, onRulesClick, onRoomsClick,
}: Props) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [bugOpen, setBugOpen] = useState(false);
  const [mobileBugOpen, setMobileBugOpen] = useState(false);
  const [bugText, setBugText] = useState("");
  const [bugStatus, setBugStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const bugBtnRef = useRef<HTMLButtonElement>(null);

  function closeBug() {
    setBugOpen(false);
    setMobileBugOpen(false);
    setBugText("");
    setBugStatus("idle");
  }

  async function submitBug() {
    if (!bugText.trim()) return;
    setBugStatus("loading");
    try {
      await sendBugReport({
        description: bugText.trim(),
        url: window.location.pathname,
        userAgent: navigator.userAgent,
      });
      setBugStatus("success");
      setBugText("");
      setTimeout(closeBug, 2000);
    } catch {
      setBugStatus("error");
    }
  }

  return (
    <header
      className="sticky top-0 z-20 border-b"
      style={{
        borderColor: "var(--d-border)",
        background: "var(--d-bg-soft)",
        backdropFilter: "blur(10px)",
        WebkitBackdropFilter: "blur(10px)",
      }}
    >
      <div className="mx-auto max-w-6xl px-4 py-3">
        <div className="flex items-center justify-between gap-2">
          {/* Logo / title */}
          <div className="flex items-center gap-2.5 min-w-0 shrink-0">
            <img src="/logo.png" alt="Дубки" className="h-11 w-11 rounded-xl object-contain shrink-0" />
            <div className="hidden md:block min-w-0">
              <h1 className="text-xl font-semibold leading-tight font-display italic truncate" style={{ color: "var(--d-text)" }}>
                Дубки — Бронирование Досуговых
              </h1>
              <div className="text-xs" style={{ color: "var(--d-text-muted)" }}>
                Общежитие ВШЭ
              </div>
            </div>
            <div className="md:hidden">
              <h1 className="text-base font-semibold leading-tight font-display italic" style={{ color: "var(--d-text)" }}>
                Дубки
              </h1>
            </div>
          </div>

          {/* Desktop actions — no flex-wrap, compact */}
          <div className="hidden sm:flex items-center gap-1.5 shrink-0">
            {/* View toggle */}
            <div className="segment-control">
              <button type="button" className={cn("segment-btn", view === "cards" && "active")} onClick={() => onToggleView("cards")}>
                Карточки
              </button>
              <button type="button" className={cn("segment-btn", view === "table" && "active")} onClick={() => onToggleView("table")}>
                Таблица
              </button>
            </div>

            {/* Rooms — icon only with tooltip */}
            <button onClick={onRoomsClick} className="btn p-2" title="Комнаты">
              <DoorOpen className="h-4 w-4" />
            </button>

            {/* Rules — text only, visible on larger screens */}
            <button onClick={onRulesClick} className="btn text-sm hidden lg:inline-flex">Правила</button>

            {user ? (
              <>
                {/* My bookings — icon only on sm/md, text on xl */}
                <button
                  onClick={onToggleMyBookings}
                  className={cn("btn p-2 xl:px-3", myBookings && "btn-active")}
                  title="Мои брони"
                  style={!myBookings && user.isAdmin ? { color: "var(--d-primary)" } : undefined}
                >
                  <BookMarked className="h-4 w-4" />
                  <span className="hidden xl:inline text-sm">Мои брони</span>
                </button>

                <button onClick={onToggleAdd} className="btn btn-primary text-sm">
                  + Бронь
                </button>

                {user.isAdmin && (
                  <button onClick={onAdminClick} className="btn p-2" title="Панель администратора">
                    <Shield className="h-4 w-4" style={{ color: "var(--d-primary)" }} />
                  </button>
                )}

                <button
                  onClick={onLogout}
                  className="btn p-2 xl:px-3"
                  title={`Выйти (${user.email})`}
                >
                  <LogOut className="h-4 w-4" />
                  <span className="hidden xl:inline text-xs max-w-[6rem] truncate" style={{ color: "var(--d-text-sec)" }}>
                    {user.email}
                  </span>
                </button>
              </>
            ) : (
              <button onClick={onLoginClick} className="btn btn-primary text-sm">
                <LogIn className="h-4 w-4" />
                Войти
              </button>
            )}

            {/* Bug report — popover instead of mailto */}
            <button
              ref={bugBtnRef}
              className="btn p-2"
              title="Сообщить об ошибке"
              onClick={() => setBugOpen(v => !v)}
            >
              <Bug className="h-4 w-4" style={{ color: "var(--d-text-muted)" }} />
            </button>
            <Popover open={bugOpen} onClose={closeBug} anchor={bugBtnRef.current} align="right">
              <div style={{ width: 260 }} className="space-y-2.5">
                <div className="font-semibold text-sm" style={{ color: "var(--d-text)" }}>Сообщить об ошибке</div>
                {bugStatus === "success" ? (
                  <p className="text-sm" style={{ color: "var(--d-text-sec)" }}>Отправлено, спасибо!</p>
                ) : (
                  <>
                    <textarea
                      className="w-full rounded-lg border px-2.5 py-2 text-sm resize-none focus:outline-none"
                      style={{
                        background: "var(--d-bg)",
                        borderColor: "var(--d-border)",
                        color: "var(--d-text)",
                        minHeight: 80,
                      }}
                      placeholder="Что пошло не так?"
                      value={bugText}
                      onChange={(e) => { setBugText(e.target.value); setBugStatus("idle"); }}
                      disabled={bugStatus === "loading"}
                    />
                    {bugStatus === "error" && (
                      <p className="text-xs" style={{ color: "var(--d-error, #e53e3e)" }}>
                        Не удалось отправить, попробуй ещё раз
                      </p>
                    )}
                    <button
                      className="btn btn-primary w-full text-sm"
                      onClick={submitBug}
                      disabled={bugStatus === "loading" || !bugText.trim()}
                    >
                      {bugStatus === "loading" ? "Отправка..." : "Отправить"}
                    </button>
                  </>
                )}
              </div>
            </Popover>
          </div>

          {/* Mobile: compact bar */}
          <div className="flex sm:hidden items-center gap-2">
            {user && (
              <button onClick={onToggleAdd} className="btn btn-primary px-3 py-2 text-sm">
                + Бронь
              </button>
            )}
            <button className="btn p-2" onClick={() => setMobileMenuOpen(o => !o)}>
              {mobileMenuOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
            </button>
          </div>
        </div>

        {/* Mobile dropdown menu — CSS grid unfurl, always in DOM */}
        <div className={cn("mobile-menu sm:hidden", mobileMenuOpen && "open")}>
          <div>
          <div className="mt-3 pb-2 space-y-2 border-t pt-3" style={{ borderColor: "var(--d-border)" }}>
            <div className="segment-control w-full">
              <button type="button" className={cn("segment-btn flex-1", view === "cards" && "active")} onClick={() => { onToggleView("cards"); setMobileMenuOpen(false); }}>
                Карточки
              </button>
              <button type="button" className={cn("segment-btn flex-1", view === "table" && "active")} onClick={() => { onToggleView("table"); setMobileMenuOpen(false); }}>
                Таблица
              </button>
            </div>

            {user && (
              <button onClick={() => { onToggleMyBookings(); setMobileMenuOpen(false); }} className={cn("btn w-full justify-start text-sm", myBookings && "btn-active")}>
                <BookMarked className="h-4 w-4" />
                Мои брони
              </button>
            )}

            <button onClick={() => { onRoomsClick(); setMobileMenuOpen(false); }} className="btn w-full justify-start text-sm">
              <DoorOpen className="h-4 w-4" />
              Комнаты
            </button>

            <button onClick={() => { onRulesClick(); setMobileMenuOpen(false); }} className="btn w-full justify-start text-sm">
              Правила
            </button>

            {user?.isAdmin && (
              <button onClick={() => { onAdminClick(); setMobileMenuOpen(false); }} className="btn w-full justify-start text-sm">
                <Shield className="h-4 w-4" style={{ color: "var(--d-primary)" }} />
                Панель администратора
              </button>
            )}

            {user ? (
              <button onClick={() => { onLogout(); setMobileMenuOpen(false); }} className="btn w-full justify-start text-sm">
                <LogOut className="h-4 w-4" />
                Выйти ({user.email})
              </button>
            ) : (
              <button onClick={() => { onLoginClick(); setMobileMenuOpen(false); }} className="btn btn-primary w-full justify-start text-sm">
                <LogIn className="h-4 w-4" />
                Войти
              </button>
            )}

            {mobileBugOpen ? (
              <div className="space-y-2.5">
                <div className="font-semibold text-sm" style={{ color: "var(--d-text)" }}>Сообщить об ошибке</div>
                {bugStatus === "success" ? (
                  <p className="text-sm" style={{ color: "var(--d-text-sec)" }}>Отправлено, спасибо!</p>
                ) : (
                  <>
                    <textarea
                      className="w-full rounded-lg border px-2.5 py-2 text-sm resize-none focus:outline-none"
                      style={{
                        background: "var(--d-bg)",
                        borderColor: "var(--d-border)",
                        color: "var(--d-text)",
                        minHeight: 80,
                      }}
                      placeholder="Что пошло не так?"
                      value={bugText}
                      onChange={(e) => { setBugText(e.target.value); setBugStatus("idle"); }}
                      disabled={bugStatus === "loading"}
                    />
                    {bugStatus === "error" && (
                      <p className="text-xs" style={{ color: "var(--d-error, #e53e3e)" }}>
                        Не удалось отправить, попробуй ещё раз
                      </p>
                    )}
                    <div className="flex gap-2">
                      <button className="btn w-full text-sm" onClick={() => { setMobileBugOpen(false); setBugText(""); setBugStatus("idle"); }}>Отмена</button>
                      <button
                        className="btn btn-primary w-full text-sm"
                        onClick={submitBug}
                        disabled={bugStatus === "loading" || !bugText.trim()}
                      >
                        {bugStatus === "loading" ? "Отправка..." : "Отправить"}
                      </button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <button
                className="btn w-full justify-start text-sm"
                onClick={() => { setMobileBugOpen(true); }}
              >
                <Bug className="h-4 w-4" />
                Сообщить об ошибке
              </button>
            )}
          </div>
          </div>
        </div>
      </div>
    </header>
  );
}
