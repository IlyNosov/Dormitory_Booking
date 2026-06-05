import { useCallback, useEffect, useState } from "react";
import type { User } from "../types/bookings";
import { getMe, logout as apiLogout } from "../api/auth";

export function useAuth() {
  const [user, setUser] = useState<User | null | undefined>(undefined); // undefined = loading

  useEffect(() => {
    getMe().then(setUser);
  }, []);

  const refresh = useCallback(async () => {
    const u = await getMe();
    setUser(u);
    return u;
  }, []);

  const logout = useCallback(async () => {
    await apiLogout();
    setUser(null);
  }, []);

  return {
    user,
    isLoading: user === undefined,
    isAuth: !!user,
    isAdmin: user?.isAdmin ?? false,
    isSuperAdmin: user?.isSuperAdmin ?? false,
    refresh,
    logout,
  };
}
