import { NextResponse } from "next/server";
import { z } from "zod";
import { setActiveUnitCookie } from "@/server/cookies";

export const dynamic = "force-dynamic";

const bodySchema = z.object({
  unit_id: z.string().min(1, "unit_id é obrigatório."),
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
      { code: "invalid_body", message: "unit_id é obrigatório." },
      { status: 400 },
    );
  }

  await setActiveUnitCookie(parsed.data.unit_id);
  return new NextResponse(null, { status: 204 });
}
