import type { User } from "./user";

export interface LoginRequest {
  email: string;
  password: string;
}

/**
 * Raw payload returned by the backend's `/auth/login`. This shape must
 * NEVER be sent to the browser: it is only ever consumed inside the Next.js
 * route handler (BFF), which reads `access_token`/`refresh_token` to set
 * httpOnly cookies and then discards them from the response body.
 */
export interface BackendLoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

/**
 * What the BFF's `/api/auth/login` route actually returns to the client.
 * Intentionally excludes both `access_token` and `refresh_token` — tokens
 * live exclusively in httpOnly cookies set server-side and are never part
 * of any JSON body or client-readable state.
 */
export interface LoginResponse {
  user: User;
}
