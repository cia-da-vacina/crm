import type { MeResponse } from "@/domain";
import { bffFetch, bffPost } from "./http";

export interface LoginResult {
  user: MeResponse;
}

export interface SessionResult {
  user: MeResponse | null;
  active_unit_id: string | null;
}

/** POST /api/auth/login — sets httpOnly session cookies server-side. */
export async function login(email: string, password: string): Promise<LoginResult> {
  return bffPost<LoginResult>("/api/auth/login", { email, password });
}

/** POST /api/auth/logout — clears the session cookies server-side. */
export async function logout(): Promise<void> {
  await bffPost<void>("/api/auth/logout");
}

/** GET /api/auth/session — current user (if any) plus the active unit id. */
export async function getSession(): Promise<SessionResult> {
  return bffFetch<SessionResult>("/api/auth/session", { cache: "no-store" });
}

/** POST /api/auth/unit — persists the active unit selection. */
export async function setActiveUnit(unitId: string): Promise<void> {
  await bffPost<void>("/api/auth/unit", { unit_id: unitId });
}

export const authService = {
  login,
  logout,
  getSession,
  setActiveUnit,
};
