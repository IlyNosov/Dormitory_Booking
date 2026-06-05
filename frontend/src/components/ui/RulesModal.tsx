import React from "react";
import { Modal } from "./Modal";
import { Clock, Shield, Users, CalendarX, Timer, Hash, UserCheck } from "lucide-react";

interface RuleRowProps {
  icon: React.ReactNode;
  text: string;
}

function RuleRow({ icon, text }: RuleRowProps) {
  return (
    <div
      className="flex items-start gap-3 rounded-xl px-3 py-2.5 border"
      style={{ borderColor: "var(--d-border)", background: "var(--d-panel)" }}
    >
      <div className="shrink-0 mt-0.5" style={{ color: "var(--d-primary)" }}>{icon}</div>
      <p className="text-sm leading-snug" style={{ color: "var(--d-text)" }}>{text}</p>
    </div>
  );
}

export function RulesModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <Modal open={open} onClose={onClose}>
      <div className="space-y-4">
        <div className="text-lg font-bold" style={{ color: "var(--d-text)" }}>
          Правила бронирования
        </div>

        <p className="text-sm" style={{ color: "var(--d-text-sec)" }}>
          Досуговые комнаты доступны для бронирования всем жильцам общежития.
          Для создания брони необходима авторизация с корпоративной почтой ВШЭ.
        </p>

        {/* Schedule */}
        <div>
          <div className="flex items-center gap-2 mb-2 font-semibold text-sm" style={{ color: "var(--d-text)" }}>
            <Clock className="h-4 w-4" style={{ color: "var(--d-primary)" }} />
            Расписание работы
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Пн–Чт: Комн. 21, 256 — 06:00–23:00" />
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Пн–Чт: Комн. 132 — 06:00–22:00" />
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Пт–Сб: Комн. 21, 256 — 06:00–01:00" />
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Пт–Сб: Комн. 132 — 06:00–23:00" />
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Вс: Комн. 21, 256 — 06:00–23:00" />
            <RuleRow icon={<Clock className="h-3.5 w-3.5" />} text="Вс: Комн. 132 — 06:00–22:00" />
          </div>
        </div>

        {/* Types */}
        <div>
          <div className="flex items-center gap-2 mb-2 font-semibold text-sm" style={{ color: "var(--d-text)" }}>
            <Users className="h-4 w-4" style={{ color: "var(--d-success)" }} />
            Типы мероприятий
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <RuleRow icon={<Users className="h-3.5 w-3.5" />} text="Публичное — открыто для всех. Любой может присоединиться." />
            <RuleRow icon={<Shield className="h-3.5 w-3.5" />} text="Частное (ЧП) — закрытое. Не приходите без приглашения." />
          </div>
        </div>

        {/* General limits */}
        <div>
          <div className="flex items-center gap-2 mb-2 font-semibold text-sm" style={{ color: "var(--d-text)" }}>
            <Timer className="h-4 w-4" style={{ color: "var(--d-secondary)" }} />
            Общие ограничения
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <RuleRow icon={<Timer className="h-3.5 w-3.5" />} text="Максимальная длительность любой брони — 4 часа." />
            <RuleRow icon={<UserCheck className="h-3.5 w-3.5" />} text="Не более 1 брони в день на человека." />
          </div>
        </div>

        {/* Private limits */}
        <div>
          <div className="flex items-center gap-2 mb-2 font-semibold text-sm" style={{ color: "var(--d-text)" }}>
            <Shield className="h-4 w-4" style={{ color: "var(--d-secondary)" }} />
            Ограничения для частных посиделок (ЧП)
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <RuleRow icon={<Shield className="h-3.5 w-3.5" />} text="ЧП разрешены в любой день, включая пятницу–воскресенье." />
            <RuleRow icon={<Hash className="h-3.5 w-3.5" />} text="Не более 3 ЧП в день на комнату." />
            <RuleRow icon={<Hash className="h-3.5 w-3.5" />} text="Не более 1 ЧП вечером (с 18:00) на комнату." />
          </div>
        </div>

        {/* Cancel */}
        <div>
          <div className="flex items-center gap-2 mb-2 font-semibold text-sm" style={{ color: "var(--d-text)" }}>
            <CalendarX className="h-4 w-4" style={{ color: "var(--d-error)" }} />
            Отмена
          </div>
          <RuleRow icon={<CalendarX className="h-3.5 w-3.5" />} text="Отменить бронь может её создатель или администратор. Не можете прийти — отмените заранее." />
        </div>

        <div className="flex justify-end pt-1">
          <button className="btn btn-primary" onClick={onClose}>Понятно</button>
        </div>
      </div>
    </Modal>
  );
}
