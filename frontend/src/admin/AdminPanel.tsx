import React, { useEffect, useState } from "react";
import { X, CalendarDays, ShieldCheck } from "lucide-react";
import { cn } from "../utils/cn";
import type { RoomInfo } from "../types/bookings";
import { fetchRooms } from "../api/rooms";
import { BookingsTab } from "./tabs/BookingsTab";
import { AdminsTab } from "./tabs/AdminsTab";

type Tab = "bookings" | "admins";

interface Props {
  open: boolean;
  onClose: () => void;
  isSuperAdmin: boolean;
  onForcePushSuccess?: () => void;
}

export function AdminPanel({ open, onClose, isSuperAdmin, onForcePushSuccess }: Props) {
  const [tab, setTab] = useState<Tab>("bookings");
  const [rooms, setRooms] = useState<RoomInfo[]>([]);

  async function loadRooms() {
    setRooms(await fetchRooms());
  }

  useEffect(() => {
    if (open) loadRooms();
  }, [open]);

  if (!open) return null;

  const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
    { key: "bookings", label: "Брони",          icon: <CalendarDays className="h-4 w-4" /> },
    { key: "admins",   label: "Администраторы",  icon: <ShieldCheck className="h-4 w-4" /> },
  ];

  return (
    <div className="fixed inset-0 z-50 flex">
      {/* Backdrop */}
      <div
        className="absolute inset-0 backdrop-blur-sm"
        style={{ background: "rgba(44,40,37,0.45)" }}
        onClick={onClose}
      />

      {/* Panel — bottom sheet on mobile, side panel on desktop */}
      <div
        className={cn(
          "relative w-full flex flex-col shadow-2xl",
          "sm:ml-auto sm:max-w-2xl sm:w-full",
          "mt-auto sm:mt-0",
          "panel-slide-in admin-panel-mobile",
        )}
        style={{
          background: "var(--d-surface)",
          borderTop: "1px solid var(--d-border)",
          borderLeft: "1px solid var(--d-border)",
          maxHeight: "90vh",
          height: "90vh",
        }}
        data-admin-panel="true"
      >
        {/* Header */}
        <div
          className="flex items-center justify-between px-5 py-4 border-b shrink-0"
          style={{ borderColor: "var(--d-border)" }}
        >
          <div className="font-bold text-lg flex items-center gap-2" style={{ color: "var(--d-text)" }}>
            <ShieldCheck className="h-5 w-5" style={{ color: "var(--d-primary)" }} />
            Панель администратора
          </div>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть">
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Tabs */}
        <div
          className="flex gap-0 px-4 pt-3 pb-0 border-b shrink-0 overflow-x-auto scrollbar-hide"
          style={{ borderColor: "var(--d-border)" }}
        >
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={cn(
                "flex items-center gap-1.5 px-4 py-2 text-sm rounded-t-xl transition-all duration-200 -mb-px border-b-2 whitespace-nowrap font-medium",
                tab === t.key
                  ? "border-b-2 text-[color:var(--d-primary)]"
                  : "border-transparent hover:text-[color:var(--d-text)]"
              )}
              style={{
                borderBottomColor: tab === t.key ? "var(--d-primary)" : "transparent",
                color: tab === t.key ? "var(--d-primary)" : "var(--d-text-sec)",
                background: tab === t.key ? "rgba(134,156,39,0.06)" : "transparent",
              }}
            >
              {t.icon}
              {t.label}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto overscroll-contain px-5 pb-5">
          {tab === "bookings" && <BookingsTab rooms={rooms} onSuccess={onForcePushSuccess} />}
          {tab === "admins"   && <AdminsTab isSuperAdmin={isSuperAdmin} />}
        </div>
      </div>
    </div>
  );
}
