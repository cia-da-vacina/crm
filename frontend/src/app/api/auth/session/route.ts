import { NextResponse } from "next/server";
import type { MeResponse } from "@/domain";
import { BackendError, backendFetch } from "@/server/backend";
import {
  clearAuthCookies,
  getAccessToken,
  getActiveUnitId,
} from "@/server/cookies";
import { refreshSession } from "@/server/refresh-session";

export const dynamic = "force-dynamic";

export interface SessionPayload {
  user: MeResponse | null;
  active_unit_id: string | null;
}

async function fetchMe(token: string): Promise<MeResponse> {
  return backendFetch<MeResponse>("/me", { token });
}

async function emptySession(): Promise<NextResponse<SessionPayload>> {
  // Drop leftover cookies so middleware doesn't keep treating the browser as
  // authenticated while AuthGuard has no user (login ↔ app redirect loop).
  await clearAuthCookies();
  return NextResponse.json({ user: null, active_unit_id: null });
}

export async function GET(): Promise<NextResponse<SessionPayload>> {
  let token = await getAccessToken();

  // Access cookie may have expired while the refresh cookie is still valid —
  // mint a new pair before giving up.
  if (!token) {
    token = await refreshSession();
    if (!token) return emptySession();
  }

  try {
    const user = await fetchMe(token);
    const activeUnitId = await getActiveUnitId();
    return NextResponse.json({ user, active_unit_id: activeUnitId });
  } catch (error) {
    const isUnauthorized = error instanceof BackendError && error.status === 401;
    if (!isUnauthorized) {
      return emptySession();
    }

    const refreshedToken = await refreshSession();
    if (!refreshedToken) {
      return emptySession();
    }

    try {
      const user = await fetchMe(refreshedToken);
      const activeUnitId = await getActiveUnitId();
      return NextResponse.json({ user, active_unit_id: activeUnitId });
    } catch {
      return emptySession();
    }
  }
}
