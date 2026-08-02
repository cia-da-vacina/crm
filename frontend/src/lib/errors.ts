/**
 * Client-side error thrown by BFF service calls (fetches to our own
 * `/api/*` route handlers). Mirrors `BackendError` from `@/server/backend`
 * but is safe to import from client components — it carries no reference
 * to the real backend URL or tokens.
 */
export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

/** Best-effort human-readable message extraction for any thrown value. */
export function toErrorMessage(
  error: unknown,
  fallback = "Ocorreu um erro inesperado. Tente novamente.",
): string {
  if (isApiError(error)) return error.message;
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}
