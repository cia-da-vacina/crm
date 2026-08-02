import { NextResponse } from "next/server";
import { refreshSession } from "@/server/refresh-session";

export const dynamic = "force-dynamic";

export async function POST(): Promise<NextResponse> {
  const token = await refreshSession();

  if (!token) {
    return NextResponse.json(
      { code: "unauthorized", message: "Sessão expirada. Faça login novamente." },
      { status: 401 },
    );
  }

  return NextResponse.json({ ok: true });
}
