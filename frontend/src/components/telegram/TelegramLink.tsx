import React, { useCallback, useEffect, useState } from "react";
import { generateLinkToken, getLinkStatus, unlinkTelegram } from "../../api/link";
import { cn } from "../../utils/cn";

type Props = {
    sessionId: string;
    onLinked?: (telegramId: string) => void;
    onUnlinked?: () => void;
};

type State =
    | { phase: "loading" }
    | { phase: "unlinked" }
    | { phase: "pending"; token: string }
    | { phase: "linked"; telegramId: string }
    | { phase: "disabled" };

export function TelegramLink({ sessionId, onLinked, onUnlinked }: Props) {
    const [state, setState] = useState<State>({ phase: "loading" });
    const [copying, setCopying] = useState(false);

    const refresh = useCallback(async () => {
        const s = await getLinkStatus(sessionId);
        if (s.botDisabled) {
            setState({ phase: "disabled" });
            return;
        }
        if (s.linked && s.telegramId) {
            setState({ phase: "linked", telegramId: s.telegramId });
            onLinked?.(s.telegramId);
        } else {
            setState({ phase: "unlinked" });
        }
    }, [sessionId, onLinked]);

    // Опрос статуса пока ожидаем подтверждения от бота
    useEffect(() => {
        refresh();
    }, [refresh]);

    useEffect(() => {
        if (state.phase !== "pending") return;
        const id = setInterval(async () => {
            const s = await getLinkStatus(sessionId);
            if (s.linked && s.telegramId) {
                setState({ phase: "linked", telegramId: s.telegramId });
                onLinked?.(s.telegramId);
            }
        }, 3000);
        return () => clearInterval(id);
    }, [state.phase, sessionId, onLinked]);

    const handleGenerate = async () => {
        const token = await generateLinkToken(sessionId);
        if (token) setState({ phase: "pending", token });
    };

    const handleUnlink = async () => {
        await unlinkTelegram(sessionId);
        setState({ phase: "unlinked" });
        onUnlinked?.();
    };

    const handleCopy = async (text: string) => {
        await navigator.clipboard.writeText(`/link ${text}`);
        setCopying(true);
        setTimeout(() => setCopying(false), 1500);
    };

    if (state.phase === "loading") {
        return <div className="text-xs text-zinc-500 animate-pulse">Загрузка...</div>;
    }

    if (state.phase === "disabled") {
        return null;
    }

    if (state.phase === "linked") {
        return (
            <div className="flex items-center gap-2">
                <span className="text-xs text-emerald-400">
                    ✅ Telegram привязан
                </span>
                <button
                    onClick={handleUnlink}
                    className="btn text-xs py-0.5 px-2"
                    title="Отвязать Telegram"
                >
                    Отвязать
                </button>
            </div>
        );
    }

    if (state.phase === "pending") {
        const cmd = `/link ${state.token}`;
        return (
            <div className="flex flex-col gap-1.5 text-xs max-w-xs">
                <span className="text-zinc-300">Отправь боту эту команду:</span>
                <div className="flex items-center gap-1">
                    <code className="bg-zinc-800 text-emerald-300 rounded px-2 py-0.5 font-mono select-all">
                        {cmd}
                    </code>
                    <button
                        onClick={() => handleCopy(state.token)}
                        className={cn("btn py-0.5 px-2", copying && "text-emerald-400")}
                    >
                        {copying ? "Скопировано!" : "Копировать"}
                    </button>
                </div>
                <span className="text-zinc-500 animate-pulse">Ожидаю подтверждения...</span>
                <button onClick={() => setState({ phase: "unlinked" })} className="text-zinc-600 hover:text-zinc-400 text-left w-fit">
                    ✕ Отмена
                </button>
            </div>
        );
    }

    // unlinked
    return (
        <button onClick={handleGenerate} className="btn text-xs">
            Привязать Telegram
        </button>
    );
}
