import { NextResponse } from "next/server";
import { z } from "zod";
import type { BackendLoginResponse, MeResponse } from "@/domain";
import { BackendError, backendFetch } from "@/server/backend";
import { setAuthCookies } from "@/server/cookies";

export const dynamic = "force-dynamic";

const bodySchema = z.object({
  email: z.string().min(1, "Email é obrigatório."),
  password: z.string().min(1, "Senha é obrigatória."),
});

export async function POST(request: Request): Promise<NextResponse> {
  let raw: unknown;
  try {
    raw = await request.json();
  } catch {
    return NextResponse.json(
      { code: "invalid_body", message: "Corpo da requisição inválido." },
      { status: 400 },
    );
  }

  const parsed = bodySchema.safeParse(raw);
  if (!parsed.success) {
    return NextResponse.json(
      { code: "invalid_body", message: "Email e senha são obrigatórios." },
      { status: 400 },
    );
  }

  try {
    const loginResult = await backendFetch<BackendLoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(parsed.data),
    });

    await setAuthCookies(
      loginResult.access_token,
      loginResult.refresh_token,
      loginResult.expires_in,
    );

    // Prefer a fresh `/me` (carries `units`); fall back to the login payload's
    // user if that call fails for some reason so login doesn't hard-fail.
    let user: MeResponse;
    try {
      user = await backendFetch<MeResponse>("/me", {
        token: loginResult.access_token,
      });
    } catch {
      user = { ...loginResult.user, units: [] };
    }

    return NextResponse.json({ user }, { status: 200 });
  } catch (error) {
    if (error instanceof BackendError) {
      return NextResponse.json(
        { code: error.code, message: error.message },
        { status: error.status || 502 },
      );
    }
    return NextResponse.json(
      { code: "unknown_error", message: "Falha inesperada ao autenticar." },
      { status: 500 },
    );
  }
}
