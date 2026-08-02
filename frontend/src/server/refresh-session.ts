import type { BackendLoginResponse } from "@/domain";
import { backendFetch } from "./backend";
import { getRefreshToken, setAuthCookies } from "./cookies";

/**
 * Attempts to mint a new access/refresh token pair from the current refresh
 * cookie and persists the result as httpOnly cookies. This is the single
 * place that knows how to talk to the backend's `/auth/refresh` endpoint —
 * every route handler that needs silent-refresh-on-401 behavior (the
 * session route and the generic proxy) goes through here so the retry logic
 * stays consistent.
 *
 * @returns the new access token on success, or `null` if there is no
 * refresh cookie or the backend rejects it (session is truly expired).
 */
export async function refreshSession(): Promise<string | null> {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) return null;

  try {
    const result = await backendFetch<BackendLoginResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    await setAuthCookies(result.access_token, result.refresh_token, result.expires_in);
    return result.access_token;
  } catch {
    return null;
  }
}
