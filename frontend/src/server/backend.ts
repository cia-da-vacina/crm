import type { ApiErrorBody } from "@/domain/api";
import { getEnv } from "./env";

/** Error thrown by `backendFetch` for any non-2xx/204 response, or network failure. */
export class BackendError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "BackendError";
    this.status = status;
    this.code = code;
  }
}

export type BackendFetchInit = RequestInit & {
  /** Bearer token to attach as `Authorization`. Pass `null`/omit for anonymous requests. */
  token?: string | null;
};

function buildUrl(basePath: string, path: string): string {
  const suffix = path.startsWith("/") ? path : `/${path}`;
  return `${basePath}${suffix}`;
}

/**
 * Server-only fetch wrapper around the backend REST API. This is the single
 * place that knows the backend's base URL and attaches bearer tokens — it
 * must only ever be called from Route Handlers, Server Actions or Server
 * Components, never from client code.
 */
export async function backendFetch<T>(
  path: string,
  init: BackendFetchInit = {},
): Promise<T> {
  const { token, headers: initHeaders, ...rest } = init;
  const env = getEnv();
  const url = buildUrl(env.API_URL, path);

  const headers = new Headers(initHeaders);
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  let res: Response;
  try {
    res = await fetch(url, {
      ...rest,
      headers,
      cache: rest.cache ?? "no-store",
    });
  } catch (cause) {
    throw new BackendError(
      0,
      "network_error",
      cause instanceof Error
        ? `Falha ao conectar ao backend: ${cause.message}`
        : "Falha ao conectar ao backend.",
    );
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const rawBody = await res.text();
  let parsed: unknown = undefined;
  if (rawBody.length > 0) {
    try {
      parsed = JSON.parse(rawBody);
    } catch {
      parsed = undefined;
    }
  }

  if (!res.ok) {
    const body = (parsed ?? {}) as Partial<ApiErrorBody>;
    throw new BackendError(
      res.status,
      body.code ?? "unknown_error",
      body.message ?? `Erro ao comunicar com o backend (status ${res.status}).`,
    );
  }

  return parsed as T;
}
