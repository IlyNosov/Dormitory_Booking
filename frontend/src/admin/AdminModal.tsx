import React, { useState } from "react";
import { Modal } from "../components/ui/Modal";
import { AlertTriangle } from "lucide-react";
import { cn } from "../utils/cn";

export function AdminModal({
                               open,
                               onClose,
                               onLogin,
                           }: {
    open: boolean;
    onClose: () => void;
    onLogin: (token: string) => boolean;
}) {
    const [value, setValue] = useState("");
    const [err, setErr] = useState("");

    function submit() {
        setErr("");
        const ok = onLogin(value);
        if (!ok) {
            setErr("Введите пароль/токен администратора.");
            return;
        }
        setValue("");
        onClose();
    }

    return (
        <Modal open={open} onClose={onClose}>
            <div className="space-y-3">
                <div className="text-lg font-semibold">Вход в админку</div>
                <div className="text-sm text-zinc-400">
                    Введите пароль (или токен). Он сохранится в браузере.
                </div>

                {err && (
                    <div className="flex items-start gap-2 rounded-xl border border-red-300/30 bg-rose-950/20 text-rose-200 px-3 py-2 text-sm">
                        <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
                        <div>{err}</div>
                    </div>
                )}

                <label className="flex flex-col gap-1 text-sm">
                    <span className="lbl">Пароль / токен</span>
                    <input
                        className={cn("field", "font-mono")}
                        value={value}
                        onChange={(e) => setValue(e.target.value)}
                        placeholder="например: admin123"
                        onKeyDown={(e) => {
                            if (e.key === "Enter") submit();
                        }}
                    />
                </label>

                <div className="flex gap-2 pt-2">
                    <button className="btn btn-primary" onClick={submit}>
                        Войти
                    </button>
                    <button className="btn" onClick={onClose}>
                        Отмена
                    </button>
                </div>
            </div>
        </Modal>
    );
}
