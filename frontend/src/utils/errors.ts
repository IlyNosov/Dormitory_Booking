export function normalizeApiError(input: unknown): string {
    const raw = typeof input === "string" ? input : (input instanceof Error ? input.message : String(input ?? ""));
    const s = raw.trim();

    try {
        const parsed = JSON.parse(s);
        const msg = parsed?.error || parsed?.message;
        if (typeof msg === "string" && msg.trim()) return translateBackendMessage(msg.trim());
    } catch {
        // not json
    }

    return translateBackendMessage(s);
}

function translateBackendMessage(msg: string): string {
    const m = msg.toLowerCase();

    if (m.includes("overlaps")) return "Эта бронь пересекается с уже существующей.";
    if (m.includes("not found")) return "Не найдено.";
    if (m.includes("forbidden")) return "Недостаточно прав для выполнения операции.";
    if (m.includes("unauthorized")) return "Нужно авторизоваться.";
    if (m.includes("invalid")) return "Некорректные данные.";
    if (m.includes("bad request")) return "Некорректный запрос.";
    if (m.includes("internal server error")) return "Ошибка сервера. Попробуйте позже.";
    if (msg.includes("<!doctype") || msg.includes("<html")) {
        return "Ошибка API: вместо JSON пришла HTML-страница. Проверь VITE_API_BASE / прокси и путь /api.";
    }

    return msg;
}
