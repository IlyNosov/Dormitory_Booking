import React, { useState } from "react";
import { Modal } from "../components/ui/Modal";
import { AlertTriangle, Mail, KeyRound } from "lucide-react";
import { cn } from "../utils/cn";
import { requestOTP, verifyOTP } from "../api/auth";
import type { User } from "../types/bookings";

type Step = "email" | "otp";

interface Props {
  open: boolean;
  onClose: () => void;
  onSuccess: (user: User) => void;
}

export function AuthModal({ open, onClose, onSuccess }: Props) {
  const [step, setStep] = useState<Step>("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function reset() {
    setStep("email");
    setEmail("");
    setCode("");
    setError("");
    setLoading(false);
  }

  function handleClose() {
    reset();
    onClose();
  }

  async function handleRequestOTP() {
    setError("");
    const trimmed = email.trim().toLowerCase();
    if (!/@(edu\.hse\.ru|hse\.ru)$/.test(trimmed)) {
      setError("Принимаются только адреса @hse.ru и @edu.hse.ru");
      return;
    }
    setLoading(true);
    try {
      await requestOTP(trimmed);
      setEmail(trimmed);
      setStep("otp");
    } catch (e: any) {
      setError(e.message || "Ошибка отправки кода");
    } finally {
      setLoading(false);
    }
  }

  async function handleVerifyOTP() {
    setError("");
    if (!/^\d{6}$/.test(code.trim())) {
      setError("Код — 6 цифр");
      return;
    }
    setLoading(true);
    try {
      const user = await verifyOTP(email, code.trim());
      reset();
      onSuccess(user);
    } catch (e: any) {
      setError(e.message || "Неверный или истёкший код");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal open={open} onClose={handleClose}>
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          {step === "email" ? (
            <Mail className="h-5 w-5" style={{ color: "var(--d-primary)" }} />
          ) : (
            <KeyRound className="h-5 w-5" style={{ color: "var(--d-primary)" }} />
          )}
          <span className="text-lg font-semibold" style={{ color: "var(--d-primary-h)" }}>
            {step === "email" ? "Вход в систему" : "Введите код"}
          </span>
        </div>

        <p className="text-sm" style={{ color: "var(--d-text-muted)" }}>
          {step === "email"
            ? "Укажите корпоративный email ВШЭ. Мы отправим вам код подтверждения."
            : `Код отправлен на ${email}. Введите его ниже (действует 10 минут).`}
        </p>

        {error && (
          <div className="banner-error">
            <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {step === "email" ? (
          <label className="flex flex-col gap-1 text-sm">
            <span className="lbl">Email</span>
            <input
              className="field"
              type="email"
              placeholder="ivanov@edu.hse.ru"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleRequestOTP()}
              autoFocus
            />
          </label>
        ) : (
          <label className="flex flex-col gap-1 text-sm">
            <span className="lbl">6-значный код</span>
            <input
              className={cn("field text-center tracking-widest text-xl font-mono")}
              type="text"
              inputMode="numeric"
              maxLength={6}
              placeholder="000000"
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
              onKeyDown={(e) => e.key === "Enter" && handleVerifyOTP()}
              autoFocus
            />
          </label>
        )}

        <div className="flex gap-2 pt-1">
          {step === "email" ? (
            <button
              className="btn btn-primary flex-1"
              onClick={handleRequestOTP}
              disabled={loading}
            >
              {loading ? "Отправка…" : "Получить код"}
            </button>
          ) : (
            <>
              <button
                className="btn btn-primary flex-1"
                onClick={handleVerifyOTP}
                disabled={loading}
              >
                {loading ? "Проверка…" : "Подтвердить"}
              </button>
              <button className="btn" onClick={() => { setStep("email"); setCode(""); setError(""); }}>
                ← Назад
              </button>
            </>
          )}
          <button className="btn" onClick={handleClose}>
            Отмена
          </button>
        </div>
      </div>
    </Modal>
  );
}
