import { NextResponse } from "next/server";
import { backendFetch } from "@/server/backend";
import { clearAuthCookies, getAccessToken } from "@/server/cookies";

export const dynamic = "force-dynamic";

export async function POST(): Promise<NextResponse> {
  const token = await getAccessToken();

  if (token) {
    try {
      await backendFetch<void>("/auth/logout", { method: "POST", token });
    } catch (error) {
      // Best-effort: local cookies are cleared below regardless of whether
      // the backend session invalidation succeeds.
      console.error("[api/auth/logout] backend logout failed", error);
    }
  }

  await clearAuthCookies();
  return new NextResponse(null, { status: 204 });
}
