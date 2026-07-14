/**
 * When mocks are on, call same-origin `/api/v1` so HTTPS pages are not
 * blocked by mixed-content rules and MSW can intercept reliably.
 */
export const API_BASE = (
  process.env.NEXT_PUBLIC_USE_MOCKS === "true"
    ? "/api/v1"
    : (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1")
).replace(/\/$/, "");
