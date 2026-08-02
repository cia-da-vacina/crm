import type { ApiErrorBody } from "@/domain/api";
import { ApiError } from "@/lib/errors";

/**
 * Client-side fetch wrapper. Every call MUST target a same-origin `/api/...`
 * route handler (the BFF) — this file must never know about, or be able to
 * reach, the real backend base URL. Auth is carried exclusively via httpOnly
 * cookies (`credentials: "same-origin"`), never via client-readable tokens.
 */
export async function bffFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  if (!path.startsWith("/api/")) {
    throw new Error(
      `[services/http] bffFetch path must start with "/api/", got: "${path}"`,
    );
  }

  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && init.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }

  let res: Response;
  try {
    res = await fetch(path, {
      ...init,
      headers,
      credentials: "same-origin",
    });
  } catch (cause) {
    throw new ApiError(
      0,
      "network_error",
      cause instanceof Error
        ? `Falha de rede: ${cause.message}`
        : "Falha de rede ao comunicar com o servidor.",
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
    throw new ApiError(
      res.status,
      body.code ?? "unknown_error",
      body.message ?? `Erro ao comunicar com o servidor (status ${res.status}).`,
    );
  }

  return parsed as T;
}

/** Query-param values that are safe to drop silently (never sent). */
type QueryValue = string | number | boolean | undefined | null;

/**
 * Builds a `?a=1&b=2` query string, skipping `undefined`/`null`/empty
 * values. Accepts any plain params object/interface — deliberately typed as
 * `object` (rather than `Record<string, QueryValue>`) so callers can pass
 * interfaces without an index signature (e.g. `ListInboxParams`).
 */
export function toQueryString(params?: object): string {
  if (!params) return "";
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params) as Array<[string, QueryValue]>) {
    if (value === undefined || value === null || value === "") continue;
    search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `?${qs}` : "";
}

export function bffGet<T>(path: string, init?: RequestInit): Promise<T> {
  return bffFetch<T>(path, { ...init, method: "GET" });
}

export function bffPost<T>(
  path: string,
  body?: unknown,
  init?: RequestInit,
): Promise<T> {
  return bffFetch<T>(path, {
    ...init,
    method: "POST",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function bffPut<T>(
  path: string,
  body?: unknown,
  init?: RequestInit,
): Promise<T> {
  return bffFetch<T>(path, {
    ...init,
    method: "PUT",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function bffPatch<T>(
  path: string,
  body?: unknown,
  init?: RequestInit,
): Promise<T> {
  return bffFetch<T>(path, {
    ...init,
    method: "PATCH",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function bffDelete<T>(path: string, init?: RequestInit): Promise<T> {
  return bffFetch<T>(path, { ...init, method: "DELETE" });
}
