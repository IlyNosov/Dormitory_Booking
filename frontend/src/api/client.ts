const API_BASE = (import.meta as any).env?.VITE_API_BASE || "";

export type ApiHeaders = {
    userEmail?: string;
    userTelegramId?: string;
};

export async function apiFetch(input: string, init: RequestInit = {}, headers?: ApiHeaders) {
    const h = new Headers(init.headers || {});
    if (!h.has("Content-Type") && init.body) h.set("Content-Type", "application/json");

    if (headers?.userEmail) h.set("X-User-Email", headers.userEmail);
    if (headers?.userTelegramId) h.set("X-User-TelegramID", headers.userTelegramId);

    const res = await fetch(`${API_BASE}${input}`, { ...init, headers: h });
    return res;
}
