const API_BASE = (import.meta as any).env?.VITE_API_BASE || "";

export type LinkStatus = {
    linked: boolean;
    telegramId?: string;
    botDisabled?: boolean;
};

function sessionHeaders(sessionId: string): Record<string, string> {
    return {
        "Content-Type": "application/json",
        "X-Session-ID": sessionId,
    };
}

export async function getLinkStatus(sessionId: string): Promise<LinkStatus> {
    const r = await fetch(`${API_BASE}/api/link/telegram`, {
        headers: sessionHeaders(sessionId),
        credentials: "include",
    });
    if (!r.ok) return { linked: false };
    return r.json();
}

export async function generateLinkToken(sessionId: string): Promise<string | null> {
    const r = await fetch(`${API_BASE}/api/link/telegram`, {
        method: "POST",
        headers: sessionHeaders(sessionId),
        credentials: "include",
    });
    if (!r.ok) return null;
    const data = await r.json();
    return data.token ?? null;
}

export async function unlinkTelegram(sessionId: string): Promise<void> {
    await fetch(`${API_BASE}/api/link/telegram`, {
        method: "DELETE",
        headers: sessionHeaders(sessionId),
        credentials: "include",
    });
}
