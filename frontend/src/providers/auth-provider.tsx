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
import type { MeResponse, Unit } from "@/domain";
import { authService } from "@/services";

export interface AuthContextValue {
  user: MeResponse | null;
  loading: boolean;
  activeUnitId: string | null;
  units: Unit[];
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setUnit: (unitId: string) => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * Session state for the whole app, backed exclusively by the BFF's
 * `/api/auth/*` route handlers. No token ever touches this component: the
 * browser only ever sees `MeResponse` (via `/api/auth/session`) and relies
 * on httpOnly cookies for actual authentication.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [activeUnitId, setActiveUnitIdState] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshSession = useCallback(async () => {
    try {
      const session = await authService.getSession();
      setUser(session.user);
      const fallbackUnitId = session.user?.units[0]?.id ?? null;
      setActiveUnitIdState(session.active_unit_id ?? fallbackUnitId);
    } catch {
      setUser(null);
      setActiveUnitIdState(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshSession();
  }, [refreshSession]);

  const login = useCallback(async (email: string, password: string) => {
    const { user: loggedInUser } = await authService.login(email, password);
    setUser(loggedInUser);

    const defaultUnitId = loggedInUser.units[0]?.id ?? null;
    if (defaultUnitId) {
      await authService.setActiveUnit(defaultUnitId);
      setActiveUnitIdState(defaultUnitId);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await authService.logout();
    } finally {
      setUser(null);
      setActiveUnitIdState(null);
    }
  }, []);

  const setUnit = useCallback(async (unitId: string) => {
    await authService.setActiveUnit(unitId);
    setActiveUnitIdState(unitId);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      loading,
      activeUnitId,
      units: user?.units ?? [],
      login,
      logout,
      setUnit,
      refreshSession,
    }),
    [user, loading, activeUnitId, login, logout, setUnit, refreshSession],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
