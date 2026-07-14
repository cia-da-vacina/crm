const ACCESS = "crm_access_token";
const REFRESH = "crm_refresh_token";
const UNIT = "crm_active_unit";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS);
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH);
}

export function setTokens(access: string, refresh: string) {
  localStorage.setItem(ACCESS, access);
  localStorage.setItem(REFRESH, refresh);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS);
  localStorage.removeItem(REFRESH);
}

export function getActiveUnitId(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(UNIT);
}

export function setActiveUnitId(unitId: string) {
  localStorage.setItem(UNIT, unitId);
}
