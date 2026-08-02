import { NextResponse } from "next/server";
import type { ApiErrorBody } from "@/domain/api";
import { getAccessToken } from "@/server/cookies";
import { getEnv } from "@/server/env";
import { refreshSession } from "@/server/refresh-session";

export const dynamic = "force-dynamic";

type RouteContext = { params: Promise<{ path: string[] }> };

const METHODS_WITH_BODY = new Set(["POST", "PUT", "PATCH", "DELETE"]);

interface BackendCallResult {
  status: number;
  body: unknown;
}

/**
 * Raw call to the backend that preserves the exact response status, so the
 * proxy can mirror it back to the client (unlike `backendFetch`, which
 * collapses any 2xx into a parsed body with no status). This is the only
 * place besides `@/server/backend` that talks to `API_URL` directly.
 */
async function callBackend(
  backendPath: string,
  search: string,
  method: string,
  requestBody: string | undefined,
  token: string | null,
): Promise<BackendCallResult> {
  const env = getEnv();
  const url = `${env.API_URL}${backendPath}${search}`;

  const headers = new Headers({ Accept: "application/json" });
  if (requestBody !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(url, {
    method,
    headers,
    body: METHODS_WITH_BODY.has(method) ? requestBody : undefined,
    cache: "no-store",
  });

  if (res.status === 204) {
    return { status: 204, body: undefined };
  }

  const raw = await res.text();
  let parsed: unknown = undefined;
  if (raw.length > 0) {
    try {
      parsed = JSON.parse(raw);
    } catch {
      parsed = undefined;
    }
  }

  return { status: res.status, body: parsed };
}

function toResponse(result: BackendCallResult): NextResponse {
  if (result.status === 204) {
    return new NextResponse(null, { status: 204 });
  }
  return NextResponse.json(result.body ?? {}, { status: result.status });
}

async function proxy(request: Request, context: RouteContext): Promise<NextResponse> {
  const { path } = await context.params;
  if (!path || path.length === 0) {
    const body: ApiErrorBody = {
      code: "not_found",
      message: "Rota de proxy inválida.",
    };
    return NextResponse.json(body, { status: 404 });
  }

  const backendPath = `/${path.join("/")}`;
  const search = new URL(request.url).search;

  let requestBody: string | undefined;
  if (METHODS_WITH_BODY.has(request.method)) {
    const raw = await request.text();
    requestBody = raw.length > 0 ? raw : undefined;
  }

  const token = await getAccessToken();

  try {
    let result = await callBackend(backendPath, search, request.method, requestBody, token);

    if (result.status === 401) {
      const refreshedToken = await refreshSession();
      if (refreshedToken) {
        result = await callBackend(
          backendPath,
          search,
          request.method,
          requestBody,
          refreshedToken,
        );
      }
    }

    return toResponse(result);
  } catch (error) {
    console.error("[api/proxy] unexpected error", error);
    const body: ApiErrorBody = {
      code: "network_error",
      message: "Falha ao conectar ao servidor.",
    };
    return NextResponse.json(body, { status: 502 });
  }
}

export async function GET(request: Request, context: RouteContext): Promise<NextResponse> {
  return proxy(request, context);
}

export async function POST(request: Request, context: RouteContext): Promise<NextResponse> {
  return proxy(request, context);
}

export async function PUT(request: Request, context: RouteContext): Promise<NextResponse> {
  return proxy(request, context);
}

export async function PATCH(request: Request, context: RouteContext): Promise<NextResponse> {
  return proxy(request, context);
}

export async function DELETE(request: Request, context: RouteContext): Promise<NextResponse> {
  return proxy(request, context);
}
