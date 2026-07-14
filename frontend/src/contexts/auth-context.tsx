"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api } from "@/lib/api";
import {
  clearTokens,
  getAccessToken,
  getActiveUnitId,
  setActiveUnitId,
  setTokens,
} from "@/lib/auth-storage";
import type { MeResponse, Unit } from "@/lib/types";

type AuthContextValue = {
  user: MeResponse | null;
  loading: boolean;
  activeUnitId: string | null;
  units: Unit[];
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setUnit: (unitId: string) => void;
  refreshMe: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeUnitId, setActiveUnit] = useState<string | null>(null);

  const refreshMe = useCallback(async () => {
    const token = getAccessToken();
    if (!token) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      const me = await api.me();
      setUser(me);
      const stored = getActiveUnitId();
      const next =
        stored && me.units.some((u) => u.id === stored)
          ? stored
          : (me.units[0]?.id ?? null);
      if (next) {
        setActiveUnitId(next);
        setActiveUnit(next);
      }
    } catch {
      clearTokens();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshMe();
  }, [refreshMe]);

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.login(email, password);
    setTokens(res.access_token, res.refresh_token);
    await refreshMe();
  }, [refreshMe]);

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } catch {
      /* ignore */
    }
    clearTokens();
    setUser(null);
  }, []);

  const setUnit = useCallback((unitId: string) => {
    setActiveUnitId(unitId);
    setActiveUnit(unitId);
  }, []);

  const value = useMemo(
    () => ({
      user,
      loading,
      activeUnitId,
      units: user?.units ?? [],
      login,
      logout,
      setUnit,
      refreshMe,
    }),
    [user, loading, activeUnitId, login, logout, setUnit, refreshMe],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
