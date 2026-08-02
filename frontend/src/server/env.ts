/**
 * Server-only environment access. Never import this from a client
 * component — it intentionally reads *unprefixed* environment variables
 * (no `NEXT_PUBLIC_`) that Next.js does not (and must not) inline into the
 * browser bundle. Meta API keys, backend base URLs and cookie policy live
 * here, never on the client.
 */

export interface ServerEnv {
  /** Base URL of the backend REST API, e.g. `https://api.ciadavacina.com.br/api/v1`. */
  API_URL: string;
  /** Whether auth cookies are marked `Secure`. Must be `true` behind HTTPS. */
  COOKIE_SECURE: boolean;
}

const DEFAULT_API_URL = "http://localhost:8080/api/v1";

let cached: ServerEnv | null = null;

function assertServerOnly(): void {
  if (typeof window !== "undefined") {
    throw new Error(
      "[server/env] getEnv() was called from the browser. Server environment " +
        "variables must never be read from client code.",
    );
  }
}

function parseBoolean(value: string | undefined, fallback: boolean): boolean {
  if (value === undefined || value === "") return fallback;
  return value === "1" || value.toLowerCase() === "true";
}

/**
 * Reads and validates server-side environment configuration. Result is
 * cached for the lifetime of the process/lambda.
 *
 * @throws {Error} if `API_URL` is missing while running in production.
 */
export function getEnv(): ServerEnv {
  assertServerOnly();

  if (cached) return cached;

  const isProduction = process.env.NODE_ENV === "production";
  const rawApiUrl = process.env.API_URL?.trim();

  if (!rawApiUrl && isProduction) {
    throw new Error(
      "[server/env] Missing required environment variable API_URL in " +
        "production. Set it to the backend base URL, e.g. " +
        "https://api.ciadavacina.com.br/api/v1.",
    );
  }

  const API_URL = (rawApiUrl && rawApiUrl.length > 0 ? rawApiUrl : DEFAULT_API_URL).replace(
    /\/+$/,
    "",
  );

  const COOKIE_SECURE = parseBoolean(process.env.COOKIE_SECURE, isProduction);

  cached = { API_URL, COOKIE_SECURE };
  return cached;
}

/** Test-only helper to reset the memoized env between test cases. */
export function __resetEnvCacheForTests(): void {
  cached = null;
}
