/**
 * Frontend-only mock API layer.
 *
 * Enable:  USE_MOCKS=true in frontend/.env.local
 * Disable: USE_MOCKS=false (or remove) and point API_URL at a real backend
 * Delete:  see README.md in this folder
 */

import { routeMockRequest, type MockRequest, type MockResponse } from "./router";

export type { MockRequest, MockResponse };

/** True when the Next.js server should answer BFF calls from in-memory fixtures. */
export function isMocksEnabled(): boolean {
  const v = process.env.USE_MOCKS?.trim().toLowerCase();
  return v === "1" || v === "true";
}

/**
 * Handle a backend-shaped request. Returns `null` when mocks are disabled
 * (callers should fall through to the real backend).
 */
export function handleMockRequest(req: MockRequest): MockResponse | null {
  if (!isMocksEnabled()) return null;
  return routeMockRequest(req);
}
