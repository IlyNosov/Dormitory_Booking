import { useCallback, useEffect, useMemo, useState } from "react";

const KEY = "adminToken";

export function useAdmin() {
    const [token, setToken] = useState<string | null>(null);

    useEffect(() => {
        setToken(localStorage.getItem(KEY));
    }, []);

    const isAdmin = !!token;

    const login = useCallback((t: string) => {
        const v = t.trim();
        if (!v) return false;
        localStorage.setItem(KEY, v);
        setToken(v);
        return true;
    }, []);

    const logout = useCallback(() => {
        localStorage.removeItem(KEY);
        setToken(null);
    }, []);

    return useMemo(
        () => ({ isAdmin, token, login, logout }),
        [isAdmin, token, login, logout],
    );
}
