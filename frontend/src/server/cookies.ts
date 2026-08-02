import { cookies } from "next/headers";
import {
  COOKIE_ACCESS,
  COOKIE_REFRESH,
  COOKIE_UNIT,
} from "./cookie-names";
import { getEnv } from "./env";

export { COOKIE_ACCESS, COOKIE_REFRESH, COOKIE_UNIT } from "./cookie-names";

/** Refresh cookie outlives the access cookie so silent refresh keeps working. */
const REFRESH_MAX_AGE_SECONDS = 60 * 60 * 24 * 30; // 30 days
const UNIT_MAX_AGE_SECONDS = 60 * 60 * 24 * 365; // 1 year

function baseCookieOptions() {
  return {
    httpOnly: true,
    sameSite: "lax" as const,
    path: "/",
    secure: getEnv().COOKIE_SECURE,
  };
}

/**
 * Persists the access/refresh tokens returned by `/auth/login` as httpOnly
 * cookies. Must only be called from a Server Action or Route Handler.
 */
export async function setAuthCookies(
  access: string,
  refresh: string,
  expiresIn: number,
): Promise<void> {
  const store = await cookies();
  const options = baseCookieOptions();

  store.set(COOKIE_ACCESS, access, {
    ...options,
    maxAge: Math.max(expiresIn, 0),
  });
  store.set(COOKIE_REFRESH, refresh, {
    ...options,
    maxAge: REFRESH_MAX_AGE_SECONDS,
  });
}

/** Clears the auth session cookies (logout). */
export async function clearAuthCookies(): Promise<void> {
  const store = await cookies();
  store.delete(COOKIE_ACCESS);
  store.delete(COOKIE_REFRESH);
}

export async function getAccessToken(): Promise<string | null> {
  const store = await cookies();
  return store.get(COOKIE_ACCESS)?.value ?? null;
}

export async function getRefreshToken(): Promise<string | null> {
  const store = await cookies();
  return store.get(COOKIE_REFRESH)?.value ?? null;
}

export async function getActiveUnitId(): Promise<string | null> {
  const store = await cookies();
  return store.get(COOKIE_UNIT)?.value ?? null;
}

/** Persists the user's active unit selection (not sensitive, but kept httpOnly for consistency). */
export async function setActiveUnitCookie(unitId: string): Promise<void> {
  const store = await cookies();
  store.set(COOKIE_UNIT, unitId, {
    ...baseCookieOptions(),
    maxAge: UNIT_MAX_AGE_SECONDS,
  });
}
